{ lib
, buildGoModule
, fetchFromGitHub
}:

buildGoModule rec {
  pname = "schemahero";
  version = "0.21.0";

  src = fetchFromGitHub {
    owner = "schemahero";
    repo = "schemahero";
    rev = "v${version}";
    hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # Update with: nix-prefetch-url --unpack https://github.com/schemahero/schemahero/archive/v${version}.tar.gz
  };

  vendorHash = "sha256-BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB="; # Update by running build and using hash from error

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

  # Rename the binary to just 'schemahero' for easier usage
  postInstall = ''
    mv $out/bin/kubectl-schemahero $out/bin/schemahero

    # Also create a symlink for kubectl plugin usage
    ln -s $out/bin/schemahero $out/bin/kubectl-schemahero
  '';

  meta = with lib; {
    description = "Cloud-native database schema management tool";
    longDescription = ''
      SchemaHero is a Kubernetes Operator for Declarative Schema Management
      for various databases. Database table schemas can be expressed as
      Kubernetes resources that can be deployed to a cluster, allowing for
      easier management and version control of database schemas.
    '';
    homepage = "https://schemahero.io";
    changelog = "https://github.com/schemahero/schemahero/releases/tag/v${version}";
    license = licenses.asl20;
    maintainers = with maintainers; [ ]; # Add your GitHub username here for nixpkgs submission
    mainProgram = "schemahero";
    platforms = platforms.unix;
  };
}
