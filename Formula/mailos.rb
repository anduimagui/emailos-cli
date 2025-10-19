class Mailos < Formula
  desc "Command-line email client powered by AI"
  homepage "https://email-os.com"
  url "https://github.com/anduimagui/emailos-cli/archiv0.1.37e/refs/tags/v0.1.37tar.gz"
  sha256 "0019dfc4b32d63c1392aa264aed2253c1e0c2fb09216f8e2cc269bbfb8bb49b5"
  license "Proprietary"
  head "https://github.com/anduimagui/emailos-cli.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.v0.1.37ersion=#{v0.1.37ersion}")
  end

  test do
    assert_match v0.1.37ersion.to_s, shell_output("#{bin}/mailos --v0.1.37ersion 2>&1")
  end
end