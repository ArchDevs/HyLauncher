import React, { useEffect, useMemo, useRef, useState } from "react";
import { ChevronDown, SquarePen, Check, Menu } from "lucide-react";
import { useTranslation } from "../i18n";

type ReleaseType = "Pre-Release" | "Release";

interface ProfileProps {
  username: string;
  currentVersion: number;
  availableVersions: number[];
  isEditing: boolean;
  onEditToggle: (val: boolean) => void;
  onUserChange: (val: string) => void;
  onVersionChange: (val: number) => void;
}

export const ProfileSection: React.FC<ProfileProps> = ({
  username,
  currentVersion,
  availableVersions,
  isEditing,
  onEditToggle,
  onUserChange,
  onVersionChange,
}) => {
  const { t } = useTranslation();
  
  // openRelease — для меню Pre-Release
  // openVersion — для анимации стрелки у vNo (и будущего меню, если захочешь)
  const [openRelease, setOpenRelease] = useState(false);
  const [openVersion, setOpenVersion] = useState(false);

  const [releaseType, setReleaseType] = useState<ReleaseType>("Pre-Release");

  const OPTIONS: ReleaseType[] = [
    t.profile.releaseType.preRelease as ReleaseType,
    t.profile.releaseType.release as ReleaseType,
  ];

  const rootRef = useRef<HTMLDivElement | null>(null);

  const baseText = "text-[#CCD9E0]/[0.90] font-[MazzardM-Medium] text-[16px]";
  const glass =
    "bg-[#090909]/[0.55] backdrop-blur-xl border border-[#7C7C7C]/[0.10]";
  const hover = "hover:bg-white/[0.04] transition";

  // close on outside click + ESC (закрываем ВСЁ синхронно)
  useEffect(() => {
    const closeAll = () => {
      setOpenRelease(false);
      setOpenVersion(false);
    };

    const onDown = (e: MouseEvent) => {
      if (!rootRef.current) return;
      if (!rootRef.current.contains(e.target as Node)) closeAll();
    };

    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") closeAll();
    };

    document.addEventListener("mousedown", onDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDown);
      document.removeEventListener("keydown", onKey);
    };
  }, []);

  const menuId = useMemo(() => "release-menu", []);

  const toggleRelease = () => {
    // чтобы открывалась только одна кнопка за раз
    setOpenRelease((v) => {
      const next = !v;
      if (next) setOpenVersion(false);
      return next;
    });
  };

  const toggleVersion = () => {
    // Only open version menu if we actually have versions
    if (availableVersions.length > 0) {
      setOpenVersion((v) => {
        const next = !v;
        if (next) setOpenRelease(false);
        return next;
      });
    }
  };

  return (
    <div className="ml-[48px]" ref={rootRef}>
      {/* Username */}
      <div
        className={`w-[280px] h-[48px] ${glass} rounded-[14px] p-4 flex items-center justify-between mb-2`}
      >
        {isEditing ? (
          <input
            autoFocus
            className={`${baseText} bg-transparent outline-none tracking-[-3%] w-full`}
            defaultValue={username}
            onBlur={(e: React.FocusEvent<HTMLInputElement>) => {
              onEditToggle(false);
              onUserChange(e.target.value);
            }}
            onKeyDown={(e: React.KeyboardEvent) =>
              e.key === "Enter" && (e.target as HTMLInputElement).blur()
            }
          />
        ) : (
          <>
            <span className={baseText}>{username}</span>
            <SquarePen
              size={16}
              className="text-[#CCD9E0]/[0.90] cursor-pointer w-[16px] h-[16px]"
              onClick={() => onEditToggle(true)}
            />
          </>
        )}
      </div>

      {/* Bottom pill */}
      <div
        className={`relative w-[280px] h-[48px] ${glass} rounded-[14px] overflow-hidden flex`}
      >
        {/* LEFT: Release type button */}
        <button
          type="button"
          aria-haspopup="menu"
          aria-expanded={openRelease}
          aria-controls={menuId}
          onClick={toggleRelease}
          className={`relative w-[132px] h-full px-[16px] flex items-center justify-between cursor-pointer ${hover}`}
        >
          <span className={`${baseText} truncate`}>{releaseType}</span>
          <ChevronDown
            size={16}
            className={`absolute right-[10px] text-[#CCD9E0]/[0.90] transition-transform ${
              openRelease ? "rotate-180" : ""
            }`}
          />
        </button>

        {/* Divider */}
        <div className="w-px h-full bg-white/10" />

        {/* MIDDLE: version button (98px) */}
        <button
          type="button"
          onClick={toggleVersion}
          className={`
            w-[98px] h-full
            pl-[16px] pr-[10px]
            flex items-center justify-between
            cursor-pointer
            ${hover}
            rounded-none
          `}
        >
          {/* текст строго 16px слева */}
          <span className={`${baseText} whitespace-nowrap`}>
            {currentVersion ? `v${currentVersion}` : `${t.profile.noVersion}`}
          </span>

          {/* стрелка работает (крутится) */}
          <ChevronDown
            size={16}
            className={`text-[#CCD9E0]/[0.90] transition-transform ${
              openVersion ? "rotate-180" : ""
            }`}
          />
        </button>

        {/* Divider */}
        <div className="w-px h-full bg-white/10" />

        {/* RIGHT: burger/menu button */}
        <button
          type="button"
          className={`
            w-[50px] h-full
            flex items-center justify-center
            cursor-pointer
            ${hover}
            rounded-none
          `}
          onClick={() => {
            // пример синхронизации: при клике закрываем остальные
            setOpenRelease(false);
            setOpenVersion(false);
            // сюда повесь открытие меню/настроек
          }}
        >
          <Menu size={16} className="text-[#CCD9E0]/[0.90]" />
        </button>

        {/* Dropdown (для Release) */}
        {openRelease && (
          <div
            id={menuId}
            role="menu"
            className="
              absolute left-0 top-[56px]
              w-[280px]
              bg-[#090909]/[0.75] backdrop-blur-[12px]
              rounded-[20px]
              border border-[#7C7C7C]/[0.10]
              overflow-hidden
              z-50
            "
          >
            {OPTIONS.map((opt, idx) => (
              <button
                key={opt}
                type="button"
                role="menuitem"
                onClick={() => {
                  setReleaseType(opt);
                  setOpenRelease(false);
                }}
                className={`
                  w-full h-[64px] px-[18px]
                  flex items-center justify-between
                  text-[#CCD9E0]/[0.90] text-[20px] font-[Mazzard]
                  hover:bg-white/[0.05]
                  cursor-pointer transition
                  ${idx !== OPTIONS.length - 1 ? "border-b border-white/10" : ""}
                `}
              >
                <span>{opt}</span>
                {opt === releaseType && <Check size={18} />}
              </button>
            ))}
          </div>
        )}

        {/* Dropdown for Versions */}
        {openVersion && availableVersions.length > 0 && (
          <div
            role="menu"
            className="
              absolute left-[132px] top-[56px]
              w-[148px]
              bg-[#090909]/[0.75] backdrop-blur-[12px]
              rounded-[20px]
              border border-[#7C7C7C]/[0.10]
              overflow-hidden
              z-50
            "
          >
            {availableVersions.map((version, idx) => (
              <button
                key={version}
                type="button"
                role="menuitem"
                onClick={() => {
                  onVersionChange(version);
                  setOpenVersion(false);
                }}
                className={`
                  w-full h-[40px] px-[18px]
                  flex items-center justify-between
                  text-[#CCD9E0]/[0.90] text-[16px] font-[Mazzard]
                  hover:bg-white/[0.05]
                  cursor-pointer transition
                  ${idx !== availableVersions.length - 1 ? "border-b border-white/10" : ""}
                `}
              >
                <span>{`v${version}`}</span>
                {version === currentVersion && <Check size={16} />}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
