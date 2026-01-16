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
'72680b088b58bb900458705e9df2d5aaa963bdcf585a979091d378ea7dc73344')

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
