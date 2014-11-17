#!/usr/bin/env bash
set -e

cd "$(dirname "$BASH_SOURCE")"

# Downloads dependencies into vendor/ directory
mkdir -p vendor
cd vendor

clone() {
	vcs=$1
	pkg=$2
	rev=$3

	pkg_url=https://$pkg
	target_dir=src/$pkg

	echo "$pkg @ $rev: "

	if [ -d $target_dir ]; then
		echo "rm old, $pkg"
		rm -fr $target_dir
	fi

	echo "clone, $pkg"
	case $vcs in
		git)
			git clone --quiet --no-checkout $pkg_url $target_dir
			( cd $target_dir && git reset --quiet --hard $rev )
			;;
		hg)
			hg clone --quiet --updaterev $rev $pkg_url $target_dir
			;;
	esac

	echo "rm VCS, $vcs"
	( cd $target_dir && rm -rf .{git,hg} )

	echo "done"
}

clone hg code.google.com/p/go.crypto aa2644fe4aa50e3b38d75187b4799b1f0c9ddcef
clone git github.com/BurntSushi/toml 2ceedfee35ad3848e49308ab0c9a4f640cfb5fb2
clone git github.com/kr/pty 67e2db24c831afa6c64fc17b4a143390674365ef

echo "don't forget to add vendor folder to your GOPATH (export GOPATH=\$GOPATH:\`pwd\`/vendor)"