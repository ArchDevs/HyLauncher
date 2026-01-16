# Maintainer: Terromur <terromuroz@proton.me>
pkgname=HyLauncher
pkgver=0.6.3
_pkgver=v0.6.3
pkgrel=2
pkgdesc="HyLauncher - unofficial Hytale Launcher for free to play gamers"
arch=('x86_64')
url="https://github.com/ArchDevs/HyLauncher"
license=('custom')
depends=('webkit2gtk' 'gtk3')
makedepends=('go' 'nodejs' 'npm')
source=("$url/archive/refs/tags/$_pkgver.tar.gz")
sha256sums=(
'4124e1675dbda6912341cd17666f732b207bcdd46eff428c3933190deb833aa8')

prepare() {
go install github.com/wailsapp/wails/v2/cmd/wails@v2.11.0
}

build() {
  cd "$pkgname-$pkgver"
  ~/go/bin/wails build
}

package() {
  cd "$pkgname-$pkgver"

  install -Dm755 "build/bin/$pkgname" "$pkgdir/usr/bin/$pkgname"

  install -Dm644 "$pkgname.desktop" "$pkgdir/usr/share/applications/$pkgname.desktop"

  install -Dm644 "$pkgname.png" "$pkgdir/usr/share/icons/hicolor/256x256/apps/$pkgname.png"
}
