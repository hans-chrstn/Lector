{
  description = "Test";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = {
    self,
    nixpkgs,
  }: let
    systems = [
      "aarch64-linux"
      "i686-linux"
      "x86_64-linux"
      "aarch64-darwin"
      "x86_64-darwin"
    ];
    forAllSystems = f:
      nixpkgs.lib.genAttrs systems (system:
        f {
          pkgs = let
            overlays = [];
          in
            import nixpkgs {
              inherit system overlays;
              config = {allowUnfree = true;};
            };
        });
  in {
    devShells = forAllSystems ({pkgs}: {
      default = pkgs.mkShell {
        packages = with pkgs; [
          go
          air
          gotools
          golangci-lint
          nodejs
          pnpm
          lua
          stylua
          prettier
          eslint
        ];
        shellHook = ''
        '';
      };
    });
  };
}
