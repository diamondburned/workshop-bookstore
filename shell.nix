{ pkgs ? import <nixpkgs> {} }:

let
	sqlc = pkgs.buildGo120Module rec {
		name = "sqlc";
		version = "1.20.0";
		src = pkgs.fetchFromGitHub {
			repo = "sqlc";
			owner = "kyleconroy";
			rev = "v${version}";
			sha256 = "sha256-ITW5jIlNoiW7sl6s5jCVRELglauZzSPmAj3PXVpdIGA=";
		};
		vendorSha256 = "sha256-5ZJPHdjg3QCB/hJ+C7oXSfzBfg0fZ+kFyMXqC7KpJmY=";
		doCheck = false;
		proxyVendor = true;
		subPackages = [ "cmd/sqlc" ];
	};
in

pkgs.mkShell {
	buildInputs = with pkgs; [
		go
		gopls
		gotools
		go-tools
		sqlc
		nodePackages.prettier
	];

	shellHook = ''
		export DB_PATH="$TMPDIR/bookstore.db"
	'';
}
