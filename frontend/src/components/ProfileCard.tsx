import React, { useEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { ChevronDown, SquarePen, Check, Menu, Loader2, AlertCircle } from "lucide-react";
import { useTranslation } from "../i18n";

export type ReleaseType = "pre-release" | "release";

interface ProfileProps {
  username: string;
  currentVersion: string;
  selectedBranch: ReleaseType;
  availableVersions: string[];
  isLoadingVersions?: boolean;
  isEditing: boolean;
  onEditToggle: (val: boolean) => void;
  onUserChange: (val: string) => void;
  onVersionChange: (val: string) => void;
  onBranchChange: (branch: ReleaseType) => void;
}

export const ProfileSection: React.FC<ProfileProps> = ({
  username,
  currentVersion,
  selectedBranch,
  availableVersions,
  isLoadingVersions = false,
  isEditing,
  onEditToggle,
  onUserChange,
  onVersionChange,
  onBranchChange,
}) => {
  const { t } = useTranslation();

  const [openRelease, setOpenRelease] = useState(false);
  const [openVersion, setOpenVersion] = useState(false);
  const [dropdownPosition, setDropdownPosition] = useState({ top: 0, left: 0 });

  const OPTIONS: { value: ReleaseType; label: string }[] = [
    { 
      value: "pre-release", 
      label: t.profile.releaseType.preRelease || "Pre-Release" 
    },
    { 
      value: "release", 
      label: t.profile.releaseType.release || "Release" 
    },
  ];

  const rootRef = useRef<HTMLDivElement | null>(null);
  const pillRef = useRef<HTMLDivElement | null>(null);
  const releaseButtonRef = useRef<HTMLButtonElement | null>(null);
  const versionButtonRef = useRef<HTMLButtonElement | null>(null);

  const baseText = "text-[#CCD9E0]/[0.90] font-[MazzardM-Medium] text-[16px]";
  const glass =
    "bg-[#090909]/[0.55] backdrop-blur-xl border border-[#7C7C7C]/[0.10]";
  const hover = "hover:bg-white/[0.04] transition";

  // Get display label for current branch
  const currentBranchLabel = useMemo(() => {
    const option = OPTIONS.find(opt => opt.value === selectedBranch);
    return option?.label || "Pre-Release";
  }, [selectedBranch, OPTIONS]);

  // Update dropdown position when opening
  useEffect(() => {
    if ((openRelease || openVersion) && pillRef.current) {
      const rect = pillRef.current.getBoundingClientRect();
      setDropdownPosition({
        top: rect.bottom + 8,
        left: rect.left,
      });
    }
  }, [openRelease, openVersion]);

  // Close on outside click + ESC
  useEffect(() => {
    const closeAll = () => {
      setOpenRelease(false);
      setOpenVersion(false);
    };

    const onDown = (e: MouseEvent) => {
      if (!rootRef.current) return;
      
      const isClickInsideRoot = rootRef.current.contains(e.target as Node);
      const isClickInsidePortal = (e.target as HTMLElement).closest('[role="menu"]');
      
      if (!isClickInsideRoot && !isClickInsidePortal) {
        closeAll();
      }
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

  const toggleRelease = (e: React.MouseEvent) => {
    e.stopPropagation();
    setOpenRelease((v) => {
      const next = !v;
      if (next) setOpenVersion(false);
      return next;
    });
  };

  const toggleVersion = (e: React.MouseEvent) => {
    e.stopPropagation();
    // Only open version menu if we have versions and not loading
    if (availableVersions.length > 0 && !isLoadingVersions) {
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
        ref={pillRef}
        className={`relative w-[280px] h-[48px] ${glass} rounded-[14px] flex`}
      >
        {/* LEFT: Release type button */}
        <button
          ref={releaseButtonRef}
          type="button"
          aria-haspopup="menu"
          aria-expanded={openRelease}
          aria-controls={menuId}
          onClick={toggleRelease}
          className={`
            relative w-[132px] h-full px-[16px] 
            flex items-center justify-between 
            rounded-l-[14px]
          `}
        >
          <span className={`${baseText} truncate`}>{currentBranchLabel}</span>
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
          ref={versionButtonRef}
          type="button"
          onClick={toggleVersion}
          className={`
            w-[98px] h-full
            pl-[16px] pr-[10px]
            flex items-center justify-between
            rounded-none
          `}
        >
          {/* Show loader while fetching versions or switching */}
          {isLoadingVersions ? (
            <>
              <span className={`${baseText} whitespace-nowrap`}>
                {isLoadingVersions ? (t.profile.loading || "Loading...") : "Switching..."}
              </span>
              <Loader2 size={16} className="text-[#CCD9E0]/[0.90] animate-spin" />
            </>
          ) : (
            <>
              <span className={`${baseText} whitespace-nowrap`}>
                {currentVersion ? (currentVersion === "auto" ? "auto" : `v${currentVersion}`) : `${t.profile.noVersion || "No Version"}`}
              </span>
              <ChevronDown
                size={16}
                className={`text-[#CCD9E0]/[0.90] transition-transform ${
                  openVersion ? "rotate-180" : ""
                }`}
              />
            </>
          )}
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
            rounded-r-[14px]
          `}
          onClick={() => {
            setOpenRelease(false);
            setOpenVersion(false);
            // Add menu/settings logic here
          }}
        >
          <Menu size={16} className="text-[#CCD9E0]/[0.90]" />
        </button>
      </div>

      {/* Dropdown for Release Branch - Using Portal */}
      {openRelease && createPortal(
        <div
          id={menuId}
          role="menu"
          style={{
            position: 'fixed',
            top: `${dropdownPosition.top}px`,
            left: `${dropdownPosition.left}px`,
            zIndex: 999999,
            pointerEvents: 'auto',
          }}
          className="
            w-[133px]
            bg-[#090909]/[0.95] backdrop-blur-[20px]
            rounded-[20px]
            border border-[#7C7C7C]/[0.20]
            overflow-hidden
            shadow-2xl
          "
        >
          {OPTIONS.map((opt, idx) => (
            <button
              key={opt.value}
              type="button"
              role="menuitem"
              onMouseDown={(e) => {
                e.preventDefault();
                e.stopPropagation();
              }}
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                onBranchChange(opt.value);
                setOpenRelease(false);
              }}
              className="
                w-full h-[64px] px-[18px]
                flex items-center justify-between
                text-[#CCD9E0]/[0.90] text-[16px] font-[Mazzard]
                hover:bg-white/[0.08]
                cursor-pointer transition-colors
                border-b border-white/10 last:border-b-0
              "
            >
              <span className="pointer-events-none">{opt.label}</span>
              {opt.value === selectedBranch && <Check size={18} className="pointer-events-none" />}
            </button>
          ))}
        </div>,
        document.body
      )}

      {/* Dropdown for Versions - Using Portal */}
      {openVersion && availableVersions.length > 0 && createPortal(
        <div
          role="menu"
          style={{
            position: 'fixed',
            top: `${dropdownPosition.top}px`,
            left: `${dropdownPosition.left + 132}px`,
            zIndex: 999999,
            pointerEvents: 'auto',
          }}
          className="
            w-[98px]
            max-h-[240px]
            overflow-y-auto
            bg-[#090909]/[0.95] backdrop-blur-[20px]
            rounded-[20px]
            border border-[#7C7C7C]/[0.20]
            shadow-2xl
            scrollbar-thin scrollbar-thumb-white/20 scrollbar-track-transparent
          "
        >
          {availableVersions.map((version, idx) => (
            <button
              key={version}
              type="button"
              role="menuitem"
              onMouseDown={(e) => {
                e.preventDefault();
                e.stopPropagation();
              }}
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                onVersionChange(version);
                setOpenVersion(false);
              }}
              className="
                w-full h-[40px] px-[18px]
                flex items-center justify-between
                text-[#CCD9E0]/[0.90] text-[16px] font-[Mazzard]
                hover:bg-white/[0.08]
                cursor-pointer transition-colors
                border-b border-white/10 last:border-b-0
              "
            >
              <span className="pointer-events-none">{version === "auto" ? "auto" : `v${version}`}</span>
              {String(version) === String(currentVersion) && <Check size={16} className="pointer-events-none" />}
            </button>
          ))}
        </div>,
        document.body
      )}
    </div>
  );
};