# Test using
# HOMEBREW_NO_INSTALL_FROM_API=1 brew audit --new helmfile-nix
# HOMEBREW_NO_INSTALL_FROM_API=1 brew install --build-from-source --verbose --debug helmfile-nix
# HOMEBREW_NO_INSTALL_FROM_API=1 brew test helmfile-nix
class HelmfileNix < Formula
  desc "Wrapper for helmfile which allows nix syntax"
  homepage "https://github.com/reMarkable/helmfile-nix"
  url "https://github.com/reMarkable/helmfile-nix.git",
    tag:      "v0.14.1",
    revision: "8dacb5847ea54df18590605ed025cd8b3e399e50"
  license "MIT"

  depends_on "go" => :build
  depends_on "helmfile"

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w")
  end

  test do
    system bin/"helmfile-nix", "--version"
  end
end
