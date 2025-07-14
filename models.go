package main

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/container/v1"
	"google.golang.org/api/sqladmin/v1beta4"
)

// ResourceType represents the type of cloud resource
type ResourceType string

const (
	ResourceTypeSQL ResourceType = "SQL"
	ResourceTypeGKE ResourceType = "GKE"
)

// CloudResource is the interface for all manageable resources
type CloudResource interface {
	GetName() string
	GetProject() string
	GetRegion() string
	GetType() ResourceType
	GetDisplayName() string
	HasPublicIP() bool
	CanAddNetwork() bool
	GetNetworkRestrictions() string
}

// SQLInstance represents a Cloud SQL instance
type SQLInstance struct {
	Name               string
	Project            string
	Region             string
	DatabaseVersion    string
	State              string
	AuthorizedNetworks []AuthorizedNetwork
	ConnectionName     string
	PublicIPEnabled    bool
	PrivateIP          string
}

func (s SQLInstance) GetName() string        { return s.Name }
func (s SQLInstance) GetProject() string     { return s.Project }
func (s SQLInstance) GetRegion() string      { return s.Region }
func (s SQLInstance) GetType() ResourceType  { return ResourceTypeSQL }
func (s SQLInstance) GetDisplayName() string { return fmt.Sprintf("%s (%s)", s.Name, s.Project) }
func (s SQLInstance) HasPublicIP() bool      { return s.PublicIPEnabled }
func (s SQLInstance) CanAddNetwork() bool    { return s.PublicIPEnabled }
func (s SQLInstance) GetNetworkRestrictions() string {
	if !s.PublicIPEnabled {
		return "Private IP only - cannot add external networks"
	}
	return ""
}

// GKECluster represents a GKE cluster
type GKECluster struct {
	Name                      string
	Project                   string
	Location                  string // Can be zone or region
	State                     string
	Endpoint                  string
	MasterAuthorizedNetworks  []AuthorizedNetwork
	PrivateClusterEnabled     bool
	PrivateEndpoint           string
	PublicEndpoint            string
}

func (g GKECluster) GetName() string        { return g.Name }
func (g GKECluster) GetProject() string     { return g.Project }
func (g GKECluster) GetRegion() string      { return g.Location }
func (g GKECluster) GetType() ResourceType  { return ResourceTypeGKE }
func (g GKECluster) GetDisplayName() string { return fmt.Sprintf("%s (%s)", g.Name, g.Project) }
func (g GKECluster) HasPublicIP() bool      { return g.PublicEndpoint != "" }
func (g GKECluster) CanAddNetwork() bool    { return true } // GKE always allows master authorized networks
func (g GKECluster) GetNetworkRestrictions() string {
	if g.PrivateClusterEnabled && g.PublicEndpoint == "" {
		return "Private cluster - access via private endpoint only"
	}
	return ""
}

// AuthorizedNetwork represents an authorized network entry
type AuthorizedNetwork struct {
	Kind        string `json:"kind,omitempty"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	DisplayName string `json:"displayName,omitempty"` // For GKE
}

// NetworkManager handles cloud resource operations
type NetworkManager struct {
	sqlService *sqladmin.Service
	gkeService *container.Service
	ctx        context.Context
}

// NewNetworkManager creates a new NetworkManager
func NewNetworkManager(ctx context.Context) (*NetworkManager, error) {
	sqlService, err := sqladmin.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL Admin service: %v", err)
	}

	gkeService, err := container.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Container service: %v", err)
	}

	return &NetworkManager{
		sqlService: sqlService,
		gkeService: gkeService,
		ctx:        ctx,
	}, nil
}

// ListProjects gets all projects accessible to the user
func (nm *NetworkManager) ListProjects() ([]string, error) {
	// Use gcloud to list all projects the user has access to
	// Remove limit to get all projects
	cmd := exec.Command("gcloud", "projects", "list", "--format=value(projectId)")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to current project if list fails
		currentCmd := exec.Command("gcloud", "config", "get-value", "project")
		currentOutput, currentErr := currentCmd.Output()
		if currentErr != nil {
			return nil, fmt.Errorf("failed to list projects and get current project: %v, %v", err, currentErr)
		}
		currentProject := strings.TrimSpace(string(currentOutput))
		if currentProject == "" {
			return nil, fmt.Errorf("no project set in gcloud config. Run 'gcloud config set project PROJECT_ID'")
		}
		return []string{currentProject}, nil
	}
	
	// Parse the output
	projectsStr := strings.TrimSpace(string(output))
	if projectsStr == "" {
		// Fallback to current project
		currentCmd := exec.Command("gcloud", "config", "get-value", "project")
		currentOutput, _ := currentCmd.Output()
		currentProject := strings.TrimSpace(string(currentOutput))
		if currentProject != "" {
			return []string{currentProject}, nil
		}
		return nil, fmt.Errorf("no projects found. Ensure you have access to at least one project")
	}
	
	projects := strings.Split(projectsStr, "\n")
	// Filter out empty strings and deduplicate
	projectMap := make(map[string]bool)
	var validProjects []string
	for _, p := range projects {
		p = strings.TrimSpace(p)
		if p != "" && !projectMap[p] {
			projectMap[p] = true
			validProjects = append(validProjects, p)
		}
	}
	
	return validProjects, nil
}

// ListAllResources gets all SQL instances and GKE clusters across projects in parallel
func (nm *NetworkManager) ListAllResources() ([]CloudResource, error) {
	projects, err := nm.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %v", err)
	}

	// Use channels to collect results
	type result struct {
		resources []CloudResource
		project   string
		err       error
	}
	
	// Limit concurrent requests to avoid API rate limits
	maxConcurrent := 20 // Increased for better performance
	if len(projects) < maxConcurrent {
		maxConcurrent = len(projects)
	}
	semaphore := make(chan struct{}, maxConcurrent)
	
	resultChan := make(chan result, len(projects)*2) // SQL + GKE per project
	var wg sync.WaitGroup
	
	// Launch goroutines for each project to get both SQL and GKE resources
	for _, project := range projects {
		// SQL instances
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			resources, err := nm.listSQLInstancesInProject(p)
			resultChan <- result{
				resources: resources,
				project:   p,
				err:       err,
			}
		}(project)

		// GKE clusters
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			resources, err := nm.listGKEClustersInProject(p)
			resultChan <- result{
				resources: resources,
				project:   p,
				err:       err,
			}
		}(project)
	}
	
	// Close the channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// Collect results
	var allResources []CloudResource
	failedProjects := make(map[string][]string) // project -> []resourceType
	
	for res := range resultChan {
		if res.err != nil {
			// Track failed projects but don't fail the entire operation
			resourceType := "unknown"
			if strings.Contains(res.err.Error(), "SQL") {
				resourceType = "SQL"
			} else if strings.Contains(res.err.Error(), "GKE") {
				resourceType = "GKE"
			}
			failedProjects[res.project] = append(failedProjects[res.project], resourceType)
			continue
		}
		allResources = append(allResources, res.resources...)
	}
	
	// Sort resources by type, project, then name
	sortResources(allResources)
	
	return allResources, nil
}

// listSQLInstancesInProject gets SQL instances from a specific project
func (nm *NetworkManager) listSQLInstancesInProject(project string) ([]CloudResource, error) {
	call := nm.sqlService.Instances.List(project)
	resp, err := call.Context(nm.ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list SQL instances in project %s: %v", project, err)
	}

	var resources []CloudResource
	for _, instance := range resp.Items {
		// Check if instance has public IP
		hasPublicIP := false
		privateIP := ""
		
		if instance.IpAddresses != nil {
			for _, ip := range instance.IpAddresses {
				if ip.Type == "PRIMARY" {
					hasPublicIP = true
				} else if ip.Type == "PRIVATE" {
					privateIP = ip.IpAddress
				}
			}
		}
		
		// Alternative check via settings
		if instance.Settings != nil && instance.Settings.IpConfiguration != nil {
			hasPublicIP = instance.Settings.IpConfiguration.Ipv4Enabled
		}

		sqlInstance := SQLInstance{
			Name:            instance.Name,
			Project:         project,
			Region:          instance.Region,
			DatabaseVersion: instance.DatabaseVersion,
			State:           instance.State,
			ConnectionName:  instance.ConnectionName,
			PublicIPEnabled: hasPublicIP,
			PrivateIP:       privateIP,
		}

		// Convert authorized networks
		if instance.Settings != nil && instance.Settings.IpConfiguration != nil {
			for _, network := range instance.Settings.IpConfiguration.AuthorizedNetworks {
				sqlInstance.AuthorizedNetworks = append(sqlInstance.AuthorizedNetworks, AuthorizedNetwork{
					Kind:  network.Kind,
					Name:  network.Name,
					Value: network.Value,
				})
			}
		}

		resources = append(resources, sqlInstance)
	}

	return resources, nil
}

// listGKEClustersInProject gets GKE clusters from a specific project
func (nm *NetworkManager) listGKEClustersInProject(project string) ([]CloudResource, error) {
	// List clusters in all locations
	parent := fmt.Sprintf("projects/%s/locations/-", project)
	call := nm.gkeService.Projects.Locations.Clusters.List(parent)
	resp, err := call.Context(nm.ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list GKE clusters in project %s: %v", project, err)
	}

	var resources []CloudResource
	for _, cluster := range resp.Clusters {
		gkeCluster := GKECluster{
			Name:     cluster.Name,
			Project:  project,
			Location: cluster.Location,
			State:    cluster.Status,
			Endpoint: cluster.Endpoint,
		}

		// Check private cluster configuration
		if cluster.PrivateClusterConfig != nil {
			gkeCluster.PrivateClusterEnabled = cluster.PrivateClusterConfig.EnablePrivateNodes
			gkeCluster.PrivateEndpoint = cluster.PrivateClusterConfig.PrivateEndpoint
			gkeCluster.PublicEndpoint = cluster.PrivateClusterConfig.PublicEndpoint
		} else {
			gkeCluster.PublicEndpoint = cluster.Endpoint
		}

		// Convert master authorized networks
		if cluster.MasterAuthorizedNetworksConfig != nil && cluster.MasterAuthorizedNetworksConfig.Enabled {
			for _, network := range cluster.MasterAuthorizedNetworksConfig.CidrBlocks {
				gkeCluster.MasterAuthorizedNetworks = append(gkeCluster.MasterAuthorizedNetworks, AuthorizedNetwork{
					Name:        network.DisplayName,
					DisplayName: network.DisplayName,
					Value:       network.CidrBlock,
				})
			}
		}

		resources = append(resources, gkeCluster)
	}

	return resources, nil
}

// GetResourceDetails fetches detailed information for a specific resource
func (nm *NetworkManager) GetResourceDetails(resource CloudResource) (CloudResource, error) {
	switch r := resource.(type) {
	case SQLInstance:
		return nm.getSQLInstanceDetails(r.Project, r.Name)
	case GKECluster:
		return nm.getGKEClusterDetails(r.Project, r.Location, r.Name)
	default:
		return nil, fmt.Errorf("unknown resource type")
	}
}

// getSQLInstanceDetails gets detailed info for a SQL instance
func (nm *NetworkManager) getSQLInstanceDetails(project, instanceName string) (CloudResource, error) {
	instance, err := nm.sqlService.Instances.Get(project, instanceName).Context(nm.ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL instance details: %v", err)
	}

	// Check if instance has public IP
	hasPublicIP := false
	privateIP := ""
	
	if instance.IpAddresses != nil {
		for _, ip := range instance.IpAddresses {
			if ip.Type == "PRIMARY" {
				hasPublicIP = true
			} else if ip.Type == "PRIVATE" {
				privateIP = ip.IpAddress
			}
		}
	}
	
	if instance.Settings != nil && instance.Settings.IpConfiguration != nil {
		hasPublicIP = instance.Settings.IpConfiguration.Ipv4Enabled
	}

	sqlInstance := SQLInstance{
		Name:            instance.Name,
		Project:         project,
		Region:          instance.Region,
		DatabaseVersion: instance.DatabaseVersion,
		State:           instance.State,
		ConnectionName:  instance.ConnectionName,
		PublicIPEnabled: hasPublicIP,
		PrivateIP:       privateIP,
	}

	// Convert authorized networks
	if instance.Settings != nil && instance.Settings.IpConfiguration != nil {
		for _, network := range instance.Settings.IpConfiguration.AuthorizedNetworks {
			sqlInstance.AuthorizedNetworks = append(sqlInstance.AuthorizedNetworks, AuthorizedNetwork{
				Kind:  network.Kind,
				Name:  network.Name,
				Value: network.Value,
			})
		}
	}

	return sqlInstance, nil
}

// getGKEClusterDetails gets detailed info for a GKE cluster
func (nm *NetworkManager) getGKEClusterDetails(project, location, clusterName string) (CloudResource, error) {
	name := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, clusterName)
	cluster, err := nm.gkeService.Projects.Locations.Clusters.Get(name).Context(nm.ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get GKE cluster details: %v", err)
	}

	gkeCluster := GKECluster{
		Name:     cluster.Name,
		Project:  project,
		Location: location,
		State:    cluster.Status,
		Endpoint: cluster.Endpoint,
	}

	// Check private cluster configuration
	if cluster.PrivateClusterConfig != nil {
		gkeCluster.PrivateClusterEnabled = cluster.PrivateClusterConfig.EnablePrivateNodes
		gkeCluster.PrivateEndpoint = cluster.PrivateClusterConfig.PrivateEndpoint
		gkeCluster.PublicEndpoint = cluster.PrivateClusterConfig.PublicEndpoint
	} else {
		gkeCluster.PublicEndpoint = cluster.Endpoint
	}

	// Convert master authorized networks
	if cluster.MasterAuthorizedNetworksConfig != nil && cluster.MasterAuthorizedNetworksConfig.Enabled {
		for _, network := range cluster.MasterAuthorizedNetworksConfig.CidrBlocks {
			gkeCluster.MasterAuthorizedNetworks = append(gkeCluster.MasterAuthorizedNetworks, AuthorizedNetwork{
				Name:        network.DisplayName,
				DisplayName: network.DisplayName,
				Value:       network.CidrBlock,
			})
		}
	}

	return gkeCluster, nil
}

// AddNetworkToResource adds an authorized network to a resource
func (nm *NetworkManager) AddNetworkToResource(resource CloudResource, networkName, networkIP string) error {
	switch r := resource.(type) {
	case SQLInstance:
		if !r.PublicIPEnabled {
			return fmt.Errorf("cannot add network to SQL instance without public IP")
		}
		return nm.addNetworkToSQLInstance(r.Project, r.Name, networkName, networkIP)
	case GKECluster:
		return nm.addNetworkToGKECluster(r.Project, r.Location, r.Name, networkName, networkIP)
	default:
		return fmt.Errorf("unknown resource type")
	}
}

// addNetworkToSQLInstance adds a network to a SQL instance
func (nm *NetworkManager) addNetworkToSQLInstance(project, instanceName, networkName, networkIP string) error {
	// Normalize the IP
	normalizedIP, err := normalizeIP(networkIP)
	if err != nil {
		return fmt.Errorf("invalid IP format: %v", err)
	}

	// Get current instance configuration
	instance, err := nm.sqlService.Instances.Get(project, instanceName).Context(nm.ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get instance: %v", err)
	}

	// Initialize settings if needed
	if instance.Settings == nil {
		instance.Settings = &sqladmin.Settings{}
	}
	if instance.Settings.IpConfiguration == nil {
		instance.Settings.IpConfiguration = &sqladmin.IpConfiguration{}
	}

	// Check if network already exists
	for _, network := range instance.Settings.IpConfiguration.AuthorizedNetworks {
		if network.Value == normalizedIP {
			if network.Name == networkName {
				return fmt.Errorf("network %s with name %s already exists", normalizedIP, networkName)
			} else {
				return fmt.Errorf("network %s already exists with name %s", normalizedIP, network.Name)
			}
		}
	}

	// Add new network
	newNetwork := &sqladmin.AclEntry{
		Kind:  "sql#aclEntry",
		Name:  networkName,
		Value: normalizedIP,
	}

	instance.Settings.IpConfiguration.AuthorizedNetworks = append(
		instance.Settings.IpConfiguration.AuthorizedNetworks,
		newNetwork,
	)

	// Update the instance
	updateRequest := &sqladmin.DatabaseInstance{
		Settings: instance.Settings,
	}

	operation, err := nm.sqlService.Instances.Patch(project, instanceName, updateRequest).Context(nm.ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to update instance: %v", err)
	}

	// Wait for operation to complete (optional, could be async)
	return nm.waitForSQLOperation(project, operation.Name, 30*time.Second)
}

// addNetworkToGKECluster adds a network to a GKE cluster's master authorized networks
func (nm *NetworkManager) addNetworkToGKECluster(project, location, clusterName, networkName, networkIP string) error {
	// Normalize the IP
	normalizedIP, err := normalizeIP(networkIP)
	if err != nil {
		return fmt.Errorf("invalid IP format: %v", err)
	}

	// Get current cluster configuration
	name := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, clusterName)
	cluster, err := nm.gkeService.Projects.Locations.Clusters.Get(name).Context(nm.ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get cluster: %v", err)
	}

	// Initialize master authorized networks if needed
	if cluster.MasterAuthorizedNetworksConfig == nil {
		cluster.MasterAuthorizedNetworksConfig = &container.MasterAuthorizedNetworksConfig{
			Enabled: true,
		}
	}

	// Check if network already exists
	for _, network := range cluster.MasterAuthorizedNetworksConfig.CidrBlocks {
		if network.CidrBlock == normalizedIP {
			if network.DisplayName == networkName {
				return fmt.Errorf("network %s with name %s already exists", normalizedIP, networkName)
			} else {
				return fmt.Errorf("network %s already exists with name %s", normalizedIP, network.DisplayName)
			}
		}
	}

	// Add new network
	newNetwork := &container.CidrBlock{
		CidrBlock:   normalizedIP,
		DisplayName: networkName,
	}

	// Create update request
	updateRequest := &container.UpdateClusterRequest{
		Update: &container.ClusterUpdate{
			DesiredMasterAuthorizedNetworksConfig: &container.MasterAuthorizedNetworksConfig{
				Enabled: true,
				CidrBlocks: append(cluster.MasterAuthorizedNetworksConfig.CidrBlocks, newNetwork),
			},
		},
	}

	// Update the cluster
	operation, err := nm.gkeService.Projects.Locations.Clusters.Update(name, updateRequest).Context(nm.ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to update cluster: %v", err)
	}

	// Wait for operation to complete (optional, could be async)
	return nm.waitForGKEOperation(project, location, operation.Name, 30*time.Second)
}

// waitForSQLOperation waits for a SQL operation to complete
func (nm *NetworkManager) waitForSQLOperation(project, operationName string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(nm.ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation timeout: %v", ctx.Err())
		case <-time.After(5 * time.Second):
			op, err := nm.sqlService.Operations.Get(project, operationName).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("failed to get operation status: %v", err)
			}

			if op.Status == "DONE" {
				if op.Error != nil && len(op.Error.Errors) > 0 {
					return fmt.Errorf("operation failed: %v", op.Error.Errors[0].Message)
				}
				return nil
			}
		}
	}
}

// waitForGKEOperation waits for a GKE operation to complete
func (nm *NetworkManager) waitForGKEOperation(project, location, operationName string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(nm.ctx, timeout)
	defer cancel()

	opName := fmt.Sprintf("projects/%s/locations/%s/operations/%s", project, location, operationName)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation timeout: %v", ctx.Err())
		case <-time.After(5 * time.Second):
			op, err := nm.gkeService.Projects.Locations.Operations.Get(opName).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("failed to get operation status: %v", err)
			}

			if op.Status == "DONE" {
				if op.Error != nil {
					return fmt.Errorf("operation failed: %v", op.Error.Message)
				}
				return nil
			}
		}
	}
}

// normalizeIP ensures the IP has a CIDR suffix
func normalizeIP(ip string) (string, error) {
	// Check if it already has a CIDR suffix
	if strings.Contains(ip, "/") {
		return ip, nil
	}
	
	// Add /32 for single IP addresses
	return ip + "/32", nil
}

// sortResources sorts resources by type, project, then name
func sortResources(resources []CloudResource) {
	sort.Slice(resources, func(i, j int) bool {
		// First sort by type
		if resources[i].GetType() != resources[j].GetType() {
			return resources[i].GetType() < resources[j].GetType()
		}
		// Then by project
		if resources[i].GetProject() != resources[j].GetProject() {
			return resources[i].GetProject() < resources[j].GetProject()
		}
		// Finally by name
		return resources[i].GetName() < resources[j].GetName()
	})
}