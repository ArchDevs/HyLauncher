import banner1Image from "../assets/images/banner1.png";
import banner2Image from "../assets/images/banner2.png";
import { useTranslation } from "../i18n";
import Banner from "./Banner";

function BannersHome() {
  const { t } = useTranslation();
  return (
    <div className="flex flex-col gap-[10px]">
      {/* Max 5 banners */}
      {/* Banner1 */}
      <Banner
        variant="compact"
        iconImage={banner1Image}
        text={`${t.banners.hynexus.text} play.hynexus.fun`}
      />
      {/* Banner2 */}
      <Banner
        variant="compact"
        iconImage={banner2Image}
        text={t.banners.nctale.text}
      />
    </div>
  );
}

export default BannersHome;
