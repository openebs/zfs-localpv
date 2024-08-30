let
  sources = import ./nix/sources.nix;
  pkgs = import sources.nixpkgs { };
  nixpkgsGinkgo = import sources.nixpkgsGinkgo { };
in
pkgs.mkShell {
  name = "scripts-shell";
  buildInputs = with pkgs; [
    chart-testing
    nixpkgs-fmt
    nixpkgsGinkgo.ginkgo
    semver-tool
    yq-go
  ];
}
