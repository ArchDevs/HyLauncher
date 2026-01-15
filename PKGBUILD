pkgname=HyLauncher
pkgver=0.3.2
_pkgver=v0.3.2
pkgrel=1
pkgdesc="HyLauncher - unofficial Hytale Launcher for free to play gamers"
arch=('x86_64')
url="https://github.com/ArchDevs/HyLauncher"
license=('custom')
depends=('webkit2gtk' 'gtk3')
source=(https://github.com/ArchDevs/$pkgname/releases/download/$_pkgver/$pkgname-linux-x64 'HyLauncher.desktop' 'HyLauncher.png')
sha256sums=(
'4686bd43e410c70dafacbdee9bc2f41cf9309bd4ffa831e268859c3c3a9b3215' 
'85f507d6d5bda0c68d9c014cac014d7649dacf9d7413c2eb5557d32ab0fa600e'
'065e5283a7e30fd654e6d18706dd1ae586f193e4698f310614a0593f62285a3f')

package() {
  install -Dm755 "$pkgname-linux-x64" "$pkgdir/usr/bin/$pkgname"

  install -Dm644 "$srcdir/$pkgname.desktop" "$pkgdir/usr/share/applications/$pkgname.desktop"

  install -Dm644 "$srcdir/$pkgname.png" "$pkgdir/usr/share/icons/hicolor/256x256/apps/$pkgname.png"
}
