import Banner from "../components/Banner";
import { useTranslation } from "../i18n";
import hynexusBigImage from "../assets/images/Hynexusbig.png";
import banner1V2Image from "../assets/images/banner1-v2.png";
import nctaleBigImage from "../assets/images/nctalebig.png";
import banner2Image from "../assets/images/banner2.png";

function ServersPage() {
  const { t } = useTranslation();

  return (
    <div className="relative h-full w-full">
      {/* Title */}
      <div
        className="
          absolute
          left-[88px]
          top-[58px]
          text-[#FFFFFF]/[0.90]
          text-[22px]
          font-[600]
          tracking-[0.04em]
          uppercase
          font-[Unbounded]
        "
      >
        {t.pages.servers}
      </div>
      <div className="absolute left-[88px] top-[100px] flex flex-wrap gap-x-[22px] gap-y-[22px]">
        <Banner
          variant="large"
          backgroundImage={hynexusBigImage}
          iconImage={banner1V2Image}
          title="HyNexus"
          description="HyNexus — это Hytale, каким он должен быть. Экономика, Кланы, PVP, PVE, ждем тебя! Сейчас!"
        />
        <Banner
          variant="large"
          backgroundImage={nctaleBigImage}
          iconImage={banner2Image}
          title="NCTale"
          description={t.banners.nctale.text}
        />
        <Banner
          variant="small"
          text="По поводу рекламы пишите нашему боту @hylauncher_bot"
        />
        <Banner
          variant="small"
          text="По поводу рекламы пишите нашему боту @hylauncher_bot"
        />
        <Banner
          variant="small"
          text="По поводу рекламы пишите нашему боту @hylauncher_bot"
        />
        <Banner
          variant="small"
          text="По поводу рекламы пишите нашему боту @hylauncher_bot"
        />
        <Banner
          variant="small"
          text="По поводу рекламы пишите нашему боту @hylauncher_bot"
        />
        <Banner
          variant="small"
          text="По поводу рекламы пишите нашему боту @hylauncher_bot"
        />
      </div>
    </div>
  );
}

export default ServersPage;
