{
  description = "Lector development and verification tools";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  outputs = { nixpkgs, ... }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in {
      devShells = forAllSystems (system:
        let pkgs = import nixpkgs { inherit system; };
        in {
          default = pkgs.mkShell {
            packages = with pkgs; [ go gotools golangci-lint nodejs gnumake pkg-config sqlite ]
              ++ pkgs.lib.optionals pkgs.stdenv.isLinux [ chromium ];
            shellHook = ''
              export GOPATH="$PWD/.cache/go"
              export GOCACHE="$PWD/.cache/go-build"
              export GOLANGCI_LINT_CACHE="$PWD/.cache/golangci-lint"
              export npm_config_cache="$PWD/.cache/npm"
              export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
              ${pkgs.lib.optionalString pkgs.stdenv.isLinux ''export PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH="${pkgs.chromium}/bin/chromium"''}
            '';
          };
        });
    };
}
