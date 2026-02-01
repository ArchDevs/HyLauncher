import banner1Image from "../assets/images/banner1.png";
import { useTranslation } from "../i18n";

function BannersHome() {
  const { t } = useTranslation();
  return (
    <div className="flex flex-col gap-[14px]">
      {/* Max 5 banners */}
      {/* Banner1 */}
      <div className="w-[400px] h-[80px] flex items-center gap-[12px] rounded-[14px] border border-[#FFA845]/10 bg-[#090909]/55 backdrop-blur-[12px] px-[10px]">
        {/* Image */}
        <img
          src={banner1Image}
          alt="Banner1"
          className="w-[60px] h-[60px] rounded-[10px]"
        />
        {/* Text */}
        <div className="flex flex-col justify-center">
          <span className="text-[14px] text-center text-[#CCD9E0]/[0.90] font-[Mazzard] tracking-[-3%]">
            {t.banners.hynexus.text}
            <span> play.hynexus.fun</span>
          </span>
        </div>
      </div>
      {/* Banner2 */}
    </div>
  );
}

export default BannersHome;
