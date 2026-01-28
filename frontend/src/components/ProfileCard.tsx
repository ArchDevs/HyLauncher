import React from "react";
import { ChevronDown, SquarePen } from "lucide-react";

interface ProfileProps {
  username: string;
  currentVersion: number;
  isEditing: boolean;
  onEditToggle: (val: boolean) => void;
  onUserChange: (val: string) => void;
}

export const ProfileSection: React.FC<ProfileProps> = ({
  username,
  currentVersion,
  isEditing,
  onEditToggle,
  onUserChange,
}) => {
  return (
    <div className="ml-[48px]">
      {/* Ник сверху */}
      <div className="w-[280px] h-[48px] bg-[#090909]/[0.55] backdrop-blur-xl rounded-[14px] border border-[#7C7C7C]/[0.10] p-4 flex items-center justify-between mb-2">
        {isEditing ? (
          <input
            autoFocus
            className="text-[#CCD9E0]/[0.90] font-[MazzardM-Medium] font-[16px] outline-none tracking-[-3%]"
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
            <span className="text-[#CCD9E0]/[0.90] font-[MazzardM-Medium] font-[16px] flex items-center justify-between">
              {username}
            </span>
            <SquarePen
              size={14}
              className="text-[#CCD9E0]/[0.90] cursor-pointer w-[16px] h-[16px]"
              onClick={() => onEditToggle(true)}
            />
          </>
        )}
      </div>

      {/* Выбор версии снизу - отдельный блок */}
      <div className="w-[280px] h-[48px] bg-[#090909]/[0.55] backdrop-blur-[12px] rounded-[14px] border border-[#7C7C7C]/[0.10] p-4 flex items-center justify-between">
        <span className="text-[#CCD9E0]/[0.90]">
          {currentVersion || "Not installed"}
        </span>
        <ChevronDown size={14} className="text-[#CCD9E0]/[0.90]" />
      </div>
    </div>
  );
};
