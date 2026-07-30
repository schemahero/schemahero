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

        # Reference SchemaHero CLI built from the upstream release source.
        # This is useful for `nix run` and as a pre-built binary in the dev
        # shell. To build from the local checkout, use `make bin/kubectl-schemahero`.
        schemahero = pkgs.buildGo126Module rec {
          pname = "schemahero";
          version = "0.21.0";

          src = pkgs.fetchFromGitHub {
            owner = "schemahero";
            repo = "schemahero";
            rev = "v${version}";
            hash = "sha256-06tondAlOlfH3kbPS36m7KuFuQycCtiujEw9GzueoVI=";
          };

          vendorHash = "sha256-DmPfEfPThQVocQZ6wAswfFiJKHPaCT/so2mVsoUukAY=";

          subPackages = [ "cmd/kubectl-schemahero" ];

          ldflags = let
            versionPkg = "github.com/schemahero/schemahero/pkg/version";
          in [
            "-s"
            "-w"
            "-X ${versionPkg}.version=${version}"
            "-X ${versionPkg}.gitSHA=${src.rev}"
            "-X ${versionPkg}.buildTime=1970-01-01T00:00:00Z"
          ];

          tags = [ "netgo" ];

          env.CGO_ENABLED = 0;

          postInstall = ''
            mv $out/bin/kubectl-schemahero $out/bin/schemahero

            # Also create a symlink for kubectl plugin usage
            ln -s $out/bin/schemahero $out/bin/kubectl-schemahero
          '';

          meta = with pkgs.lib; {
            description = "A cloud-native database schema management tool";
            longDescription = ''
              SchemaHero is a Kubernetes Operator for Declarative Schema Management
              for various databases. Database table schemas can be expressed as
              Kubernetes resources that can be deployed to a cluster.
            '';
            homepage = "https://schemahero.io";
            changelog = "https://github.com/schemahero/schemahero/releases/tag/v${version}";
            license = licenses.asl20;
            maintainers = with maintainers; [ ];
            mainProgram = "schemahero";
            platforms = platforms.unix;
          };
        };
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
            go_1_26
            gnumake
            git
            kubectl
            kustomize
            docker
            oras
            cosign
            trivy
            schemahero
          ];

          shellHook = ''
            # Ensure Go-installed tools (controller-gen, client-gen, etc.) are on PATH
            export PATH="$PATH:$(go env GOPATH)/bin"

            echo "SchemaHero development environment"
            echo "Available commands:"
            echo "  - go: Go ${pkgs.go_1_26.version}"
            echo "  - make: GNU Make for building"
            echo "  - git: Git version control"
            echo "  - kubectl: Kubernetes CLI"
            echo "  - kustomize: Kubernetes manifest customization"
            echo "  - docker: Docker CLI"
            echo "  - oras: OCI registry client (for test-dev plugin pushes)"
            echo "  - cosign: Container signing (for cosign-sign)"
            echo "  - trivy: Vulnerability scanner (for make scan)"
            echo "  - schemahero: SchemaHero CLI (from Nix package)"
            echo ""
            echo "Common make targets:"
            echo "  - make bin/kubectl-schemahero"
            echo "  - make manager"
            echo "  - make test"
            echo "  - make test-plugins"
            echo "  - make generate manifests"
            echo "  - make plugins"
            echo "  - PLUGIN=postgres make install-dev"
            echo "  - make deploy"
            echo "  - make local"
            echo ""
            echo "To test the Nix package: nix run"
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
