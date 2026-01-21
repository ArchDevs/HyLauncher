# Maintainer: Terromur <terromuroz@proton.me>
pkgname=HyLauncher
pkgver=0.6.6
_pkgver=v0.6.6
pkgrel=1
pkgdesc="HyLauncher - unofficial Hytale Launcher for free to play gamers"
arch=('x86_64')
url="https://github.com/ArchDevs/HyLauncher"
license=('GPL3')
depends=('webkit2gtk' 'gtk3')
makedepends=('go' 'nodejs' 'npm')
source=("$url/archive/refs/tags/$_pkgver.tar.gz")
sha256sums=(
'759b239de86020c59c4ecd82dd2234254f8b61a2b7d2df159b99b7664f9392eb')

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
  
  install -Dm644 "LICENSE" "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
}
