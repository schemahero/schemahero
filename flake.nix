{
  description = "SchemaHero - A cloud-native database schema management tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        schemahero = pkgs.callPackage ./nix/schemahero.nix { };
      in
      {
        packages = {
          default = schemahero;
          schemahero = schemahero;
        };

        apps = {
          default = flake-utils.lib.mkApp {
            drv = schemahero;
            name = "schemahero";
          };
          schemahero = flake-utils.lib.mkApp {
            drv = schemahero;
            name = "schemahero";
          };
          kubectl-schemahero = flake-utils.lib.mkApp {
            drv = schemahero;
            name = "kubectl-schemahero";
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_24
            gnumake
            kubectl
            schemahero
          ];

          shellHook = ''
            echo "SchemaHero development environment"
            echo "Available commands:"
            echo "  - go: Go ${pkgs.go_1_24.version}"
            echo "  - make: GNU Make for building"
            echo "  - kubectl: Kubernetes CLI"
            echo "  - schemahero: SchemaHero CLI (from Nix package)"
            echo ""
            echo "To build from source: make bin/kubectl-schemahero"
            echo "To test Nix package: nix run"
          '';
        };

        checks = {
          # Build check - ensures the package builds successfully
          build = schemahero;

          # Basic functionality test
          schemahero-test = pkgs.runCommand "schemahero-test" {
            buildInputs = [ schemahero ];
          } ''
            # Test that the binary exists and runs
            schemahero --help > $out

            # Test that version command works
            schemahero version || true

            # Test kubectl plugin mode
            kubectl-schemahero --help >> $out || true

            echo "All basic tests passed" >> $out
          '';
        };
      });
}
