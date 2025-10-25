class Mailos < Formula
  desc "Command-line email client powered by AI"
  homepage "https://email-os.com"
  url "https://github.com/anduimagui/emailos-cli/archive/refs/tags/v0.1.67.tar.gz"
  sha256 "c002a8b28ddd4ec98341571126658680bb578ccd56c7bbf4613dba2c30483248"
  license "Proprietary"
  head "https://github.com/anduimagui/emailos-cli.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}"), "./cmd/mailos"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/mailos --version 2>&1")
  end
end