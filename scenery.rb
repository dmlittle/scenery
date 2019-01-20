class Scenery < Formula
  desc "Terraform plan output prettifier"
  homepage "https://github.com/dmlittle/scenery"
  url "https://github.com/dmlittle/scenery/archive/v0.1.0.tar.gz"
  sha256 "773372ac325ae746b95f0d503b08461bfa039bf9a0be6a3db2805aec69b61f74"

  depends_on "go" => :build

  def install
    ENV["GOPATH"] = buildpath

    system "go", "get", "-u", "github.com/dmlittle/scenery"
    system "go", "build", "-o", "scenery"
    bin.install "scenery"
  end

  test do
    system "#{bin}/scenery", "--version"
  end
end
