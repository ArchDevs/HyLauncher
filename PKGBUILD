pkgname=HyLauncher
pkgver=0.2
_pkgver=v0.2
pkgrel=1
pkgdesc="HyLauncher - unofficial Hytale Launcher for free to play gamers"
arch=('x86_64')
url="https://github.com/ArchDevs/HyLauncher"
license=('custom')
depends=('webkit2gtk' 'gtk3')
source=(https://github.com/ArchDevs/$pkgname/releases/download/$_pkgver/HyLauncher-linux-x64 'HyLauncher.desktop' 'HyLauncher.png')
sha256sums=(
'f46b8b7c5d7e8eb1ec6bd43aeb7d501f980a92d04ea4b47e3fe2edcb1297d850' 
'85f507d6d5bda0c68d9c014cac014d7649dacf9d7413c2eb5557d32ab0fa600e'
'065e5283a7e30fd654e6d18706dd1ae586f193e4698f310614a0593f62285a3f')

package() {
  install -Dm755 "HyLauncher-linux-x64" "$pkgdir/usr/bin/$pkgname"

  install -Dm644 "$srcdir/HyLauncher.desktop" "$pkgdir/usr/share/applications/HyLauncher.desktop"

  install -Dm644 "$srcdir/HyLauncher.png" "$pkgdir/usr/share/icons/hicolor/256x256/apps/HyLauncher.png"
}
