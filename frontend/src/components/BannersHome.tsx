import banner1Image from "../assets/images/banner1.png";
import { useTranslation } from "../i18n";
import Banner from "./Banner";

function BannersHome() {
  const { t } = useTranslation();
  return (
    <div className="flex flex-col gap-[14px]">
      {/* Max 5 banners */}
      {/* Banner1 */}
      <Banner
        variant="compact"
        iconImage={banner1Image}
        text={`${t.banners.hynexus.text} play.hynexus.fun`}
      />
      {/* Banner2 */}
    </div>
  );
}

export default BannersHome;
