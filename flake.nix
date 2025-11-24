{
  description = "ict-web";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";
  };

  outputs = { self, nixpkgs, ... }: let
    system = "x86_64-linux";
    pkgs = import nixpkgs { inherit system; };
  in {
    devShells.${system}.default = pkgs.mkShell {
      packages = with pkgs; [
        go            # Go compiler
        postgresql    # Needed for libpq client library
      ];

      shellHook = ''
        export LD_LIBRARY_PATH=${pkgs.postgresql.lib}/lib:$LD_LIBRARY_PATH
        echo "🔧 Go version: $(go version)"
        echo "🐘 libpq is available via postgresql"
        exec zsh
      '';
    };
  };
}
