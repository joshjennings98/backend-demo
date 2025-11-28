{
  description = "A flake for github.com/joshjennings98/backend-demo";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }: flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        inherit system;
      };
    in
    {
      packages = {
        default = pkgs.buildGoModule {
          pname = "backend-demo";
          version = "2.3.0";
          nativeBuildInputs = [ pkgs.pkg-config ];
          vendorHash = "sha256-lL14bXiQqxIZYIrZC7PxmseWKRYNB8MJ+x33lHMOflA=";
          src = ./backend-demo;
          meta = {
            description = "Demonstrate backend projects with the power of Go and HTMX";
          };
        };
      };
    }
  );
}
