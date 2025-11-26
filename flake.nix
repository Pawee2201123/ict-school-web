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

        export PGHOST=127.0.0.1
        export PGPORT=5432
        export PGUSER=postgres
        export PGDATABASE=ict 
        export PGSSLMODE=disable
        export ADMIN_EMAIL=admin@admin.com
        export ADMIN_PASSWORD=admin1234
        echo "🔧 Go version: $(go version)"
        echo "🐘 libpq is available via postgresql"
        exec zsh
      '';
      # TODO use secret manager for admin, cookie hash,encrypt key
    };
  };
}
