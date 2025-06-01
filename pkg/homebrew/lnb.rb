class Lnb < Formula
  desc "A cross-platform utility that makes command-line tools accessible from anywhere by creating symbolic links or wrapper scripts in your system's PATH"
  homepage "https://github.com/muthuishere/lnb"
  version "0.1.0"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/muthuishere/lnb/releases/download/v0.1.0/lnb-darwin-arm64.zip"
    sha256 "SHA256 hash will be added during release"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/muthuishere/lnb/releases/download/v0.1.0/lnb-darwin-amd64.zip"
    sha256 "SHA256 hash will be added during release"
  end

  def install
    bin.install "lnb"
  end

  test do
    system "#{bin}/lnb", "--version"
  end
end
