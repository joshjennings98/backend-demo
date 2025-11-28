# for an intro to nix syntax see https://nixos.org/guides/nix-pills/04-basics-of-language and for information on flakes see https://nixos.wiki/wiki/Flakes#Flake_schema
{
  description = "A flake for github.com/joshjennings98/backend-demo";

  inputs = { # inputs specify the dependencies of the flake
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:tweag/gomod2nix";
      # using follows in Nix flakes ensures that the input uses the same version of nixpkgs and flake-utils on your machine, maintaining compatibility and consistency across dependencies
      inputs.nixpkgs.follows = "nixpkgs"; # particularly with nixpkgs, the use of follows will ensure that the nixpkgs will align with the other inputs and not create duplicate sets of the same dependency
      inputs.utils.follows = "utils";
    };
  };

  # outputs define the build results, packages, development environments, or configurations that the flake provides, based on its inputs
  outputs = { self, nixpkgs, utils, gomod2nix }: utils.lib.eachDefaultSystem (system: # utils.lib.eachDefaultSystem is a helper function from flake-utils that iterates over all default nixos systems (like x86_64-linux, aarch64-linux, etc.) to generate outputs for each system and cross-platform compatibility
    let
      pkgs = import nixpkgs {
        inherit system;
        # overlays allow you to extend or customize nix packages by modifying or adding packages to the existing nixpkgs set without altering the original source
        overlays = [ gomod2nix.overlays.default ];
      };
    in
    {
      packages = { # this defines a set of packages in the flake, with default being the primary package that is built
        default = pkgs.buildGoApplication { # this function is used to build a Go application using gomod2nix to instead of relying on a vendor hash
          name = "backend-demo";
          src = ./backend-demo;
          pwd = ./.;
          modules = ./gomod2nix.toml; # to understand why this is used see https://www.tweag.io/blog/2021-03-04-gomod2nix/
        };
      };

      devShells.default = pkgs.mkShell { # this defines a shell that includes all the necessary tools and dependencies for development, start it using `nix develop`
        buildInputs = with pkgs; [ # buildInputs are the depenedencies of an environment or a build function
          go                                   # obviously you need golang itself
          gopls                                # language server for golang
          golangci-lint                        # tool for linting go code
          golangci-lint-langserver             # language server for golangci-lint
          gomod2nix.packages.${system}.default # make the gomod2nix available in the shell
        ];
      };
    }
  );
}
