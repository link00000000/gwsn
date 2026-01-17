{
  description = "Go Template";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        commonTools = with pkgs; [
          go
          gopls
          delve
          gnumake
          pkg-config
        ];

        mkWindowsShell =
          { crossPkgs
          , goarch
          }:
          let
            ccPrefix = crossPkgs.stdenv.cc.targetPrefix;
          in
          pkgs.mkShellNoCC {
            nativeBuildInputs =
              commonTools
              ++ [
                crossPkgs.stdenv.cc
                crossPkgs.binutils
              ];

            shellHook = ''
              export GOOS=windows
              export GOARCH=${goarch}
              export CGO_ENABLED=1
              export CC=${ccPrefix}gcc
              export CXX=${ccPrefix}g++
            '';
          };
      in
      {
        devShells.default = pkgs.mkShell {
          nativeBuildInputs =
		    commonTools
		    ++ [
			  pkgs.libayatana-appindicator
			  pkgs.gtk3
		    ];
        };

        devShells.windows-amd64 = mkWindowsShell {
          crossPkgs = pkgs.pkgsCross.mingwW64;
          goarch = "amd64";
        };

        devShells.windows-386 = mkWindowsShell {
          crossPkgs = pkgs.pkgsCross.mingw32;
          goarch = "386";
        };
      }
    );
}
