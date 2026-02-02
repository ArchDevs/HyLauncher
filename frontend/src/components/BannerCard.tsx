function BannerCard() {
  return (
    <button className="absolute left-[88px] top-[100px] w-[448px] h-[200px] rounded-[20px] border border-[#7C7C7C]/10 overflow-hidden group cursor-pointer transform-gpu">
      {/* Background image */}
      <img
        src="src/assets/images/Hynexusbig.png"
        alt=""
        className="w-full h-full object-cover opacity-90 transition-all duration-300 filter saturate-[0.6] contrast-[0.85] brightness-[0.93] group-hover:saturate-100 group-hover:contrast-100 group-hover:brightness-100 will-change-[filter]"
      />

      {/* Dark overlay */}
      <div className="absolute inset-0 bg-[#090909]/25 pointer-events-none" />

      {/* Small icon */}
      <img
        src="src/assets/images/banner1-v2.png"
        alt="Banner icon"
        className="absolute bottom-[10px] left-[10px] w-[60px] h-[60px] rounded-[10px] pointer-events-none transform-gpu"
      />

      {/* Text block */}
      <div
        className="
          absolute bottom-[14px]
          left-[80px]
          right-[14px]
          w-[310px]
          flex flex-col
          pointer-events-none
        "
      >
        {/* Title */}
        <div className="font-[Unbounded] text-[14px] text-white/[0.90] text-left">
          HyNexus
        </div>

        {/* Description */}
        <p className="mt-[2px] text-[14px] leading-[16px] font-[Mazzard] text-white/[0.85] text-justify">
          HyNexus — это Hytale, каким он должен быть. Экономика, Кланы, PVP,
          PVE, ждем тебя! Сейчас!
        </p>
      </div>
    </button>
  );
}

export default BannerCard;
