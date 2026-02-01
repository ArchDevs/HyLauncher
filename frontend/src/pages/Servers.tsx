import { useTranslation } from "../i18n";

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
    </div>
  );
}

export default ServersPage;
