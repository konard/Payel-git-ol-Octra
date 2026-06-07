{
  description = "Octra orchestrator — multi-agent AI software factory";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        # Build with buildGoModule; agents source needed for go.mod replace directive
        orchestrator = pkgs.buildGoModule {
          name = "octra-orchestrator";
          src = ./.;

          prePatch = ''
            # agents/ must be at ../agents to satisfy go.mod replace directive
            ln -sf ${builtins.path { path = ./../agents; name = "agents-source"; }} ../agents
          '';

          # Set after first build attempt:
          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
          CGO_ENABLED = 0;
          doCheck = false;
          ldflags = [ "-s" "-w" ];
          subPackages = [ "./cmd/app" ];
        };

        dockerImage = pkgs.dockerTools.buildImage {
          name = "octra-orchestrator";
          tag = "latest";
          copyToRoot = pkgs.buildEnv {
            name = "image-root";
            paths = with pkgs; [ nix git cacert orchestrator ];
            pathsToLink = [ "/bin" ];
          };
          config = {
            Env = [
              "NIX_CONFIG=experimental-features = nix-command flakes"
              "GIT_USER_NAME=CrewAI Bot"
              "GIT_USER_EMAIL=bot@crewai.local"
            ];
            Cmd = [ "${orchestrator}/bin/app" ];
          };
        };
      in {
        packages = {
          default = orchestrator;
          inherit dockerImage;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [ go_1_25 gopls gotools git nix ];
          shellHook = ''
            echo "Octra orchestrator dev shell"
            echo "  go build ./cmd/app  — build the binary"
          '';
        };
      })
    // {
      # Expose NixOS module for deployment
      nixosModules.octra-orchestrator = import ./nix/module.nix;
    };
}
