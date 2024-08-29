{
  description = "A flake for github.com/joshjennings98/backend-demo";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:tweag/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.utils.follows = "utils";
    };
  };

  outputs = { self, nixpkgs, utils, gomod2nix }: utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [ 
          gomod2nix.overlays.default 
            (final: pre: {
              update = final.writeScriptBin "update" ''
                #!/usr/bin/env bash
                ROOT_DIR=$(pwd)
                while [ ! -f "$ROOT_DIR/flake.nix" ] && [ "$ROOT_DIR" != "/" ]; do ROOT_DIR=$(dirname "$ROOT_DIR"); done # will work from anywhere IN project
                gomod2nix --dir "$ROOT_DIR/backend-demo" --outdir "$ROOT_DIR"
              '';
            })
          ];
      };
    in
    {
      packages = {
        default = pkgs.buildGoApplication {
          name = "backend-demo";
          src = ./backend-demo;
          pwd = ./.;
          modules = ./gomod2nix.toml;
          meta = {
            description = "Demonstrate backend projects with the power of Go and HTMX";
          };
        };
      };

      devShells.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          go
          gopls
          gotools
          go-tools
          gomod2nix.packages.${system}.default
          update
        ];
      };
    }
  );
}
