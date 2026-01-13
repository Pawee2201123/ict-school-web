{
  description = "ict-web";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";
  };

  outputs = { self, nixpkgs, ... }: let
    system = "x86_64-linux";
    pkgs = import nixpkgs { inherit system; };
    
    # 1. DEFINE THE APP BUILD
    # This compiles Go and copies the 'web' folder so the binary can find templates
    app = pkgs.buildGoModule {
      pname = "ict-web";
      version = "0.1.0";
      src = ./.;

      # IMPORTANT: Nix needs to lock your Go dependencies. 
      # Step 1: Set this to lib.fakeHash
      # Step 2: Run 'nix build', it will fail and show you the real hash.
      # Step 3: Copy the real hash here.
      vendorHash = "sha256-JewAZFfsD4yI/5u9u53FJJkcUCLjO3VYcaBLUFniPko="; 

      # Disable CGO to create a static binary (easier for Docker)
      CGO_ENABLED = 0;

      # After building the binary, copy the web/ templates next to it
      postInstall = ''
        mkdir -p $out/web
        cp -r web $out/
      '';
    };

    # 2. DEFINE THE DOCKER IMAGE
    dockerImage = pkgs.dockerTools.buildImage {
      name = "ict-web-image";
      tag = "latest";
      
      # We put the App, SSL Certs (for AWS S3/HTTPS), and Timezone data in the image
      copyToRoot = pkgs.buildEnv {
        name = "image-root";
        paths = [ app pkgs.cacert pkgs.tzdata ];
        pathsToLink = [ "/bin" "/web" ];
      };

      config = {
        # When container starts, run this command
        Cmd = [ "/bin/server" ]; # <--- Assuming your binary is named 'server' based on go.mod?
        # If your go build -o is "main", change this to "/bin/main"

        # Expose port 8080
        ExposedPorts = {
          "8080/tcp" = {};
        };
        
        # Set working directory so the app finds "web/templates"
        WorkingDir = "${app}"; 
        
        # Default Env Vars
        Env = [
          "SSL_CERT_FILE=${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt"
        ];
      };
    };

  in {
    # YOUR EXISTING DEV SHELL
    devShells.${system}.default = pkgs.mkShell {
      packages = with pkgs; [ go postgresql ];
      shellHook = ''
        export LD_LIBRARY_PATH=${pkgs.postgresql.lib}/lib:$LD_LIBRARY_PATH
        export PGHOST=127.0.0.1
        export PGPORT=5432
        export PGUSER=postgres
        export PGDATABASE=ict 
        export PGSSLMODE=disable
        echo "🔧 Go version: $(go version)"
        echo "🐘 libpq is available"
        exec zsh
      '';
    };

    # NEW OUTPUTS
    packages.${system} = {
      default = app;           # Run 'nix build' to build just the binary
      docker = dockerImage;    # Run 'nix build .#docker' to build the image
    };
  };
}
