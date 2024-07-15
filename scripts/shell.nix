with (import <nixpkgs> { });
mkShell {
  name = "scripts-shell";
  buildInputs = [
    semver-tool
    yq-go
    chart-testing
  ];
}
