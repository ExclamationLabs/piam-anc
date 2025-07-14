class PiamAnc < Formula
  desc "Beautiful TUI for managing Google Cloud SQL and GKE authorized networks"
  homepage "https://github.com/ExclamationLabs/piam-anc"
  version "1.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/ExclamationLabs/piam-anc/releases/download/v1.0.0/piam-anc-darwin-amd64.tar.gz"
      sha256 "2060c4594812f245e4c52e9536ce20ffd52c1c23827265060f74437d193fc0f1"
    else
      url "https://github.com/ExclamationLabs/piam-anc/releases/download/v1.0.0/piam-anc-darwin-arm64.tar.gz"
      sha256 "9a3208e3e71b9e55e7cfaa07a0e73b3e73f01b953a82a0e29129d76a17c382ee"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/ExclamationLabs/piam-anc/releases/download/v1.0.0/piam-anc-linux-amd64.tar.gz"
      sha256 "3688e4e8db9fe6f2b9cb8c7e564c0b5bc02e78b3af818d6a9ec807ca37e5d3fe"
    else
      url "https://github.com/ExclamationLabs/piam-anc/releases/download/v1.0.0/piam-anc-linux-arm64.tar.gz"
      sha256 "e8d37462d105cc6721d6bef009fd7e2de8ee1b78285d20075236ad1df0985c9b"
    end
  end

  def install
    bin.install "piam-anc"
  end

  def post_install
    # Create completion scripts directory
    (var/"piam-anc").mkpath
  end

  def caveats
    <<~EOS
      PIAM Admin Network Configurator has been installed!

      Before using piam-anc, ensure you're authenticated with Google Cloud:
        gcloud auth application-default login

      To get started:
        piam-anc

      For help:
        piam-anc --help

      Required GCP permissions:
        - cloudsql.instances.list/get/update
        - container.clusters.list/get/update  
        - resourcemanager.projects.list
    EOS
  end

  test do
    assert_match "PIAM Admin Network Configurator", shell_output("#{bin}/piam-anc --version")
  end
end