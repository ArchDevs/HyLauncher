import React from "react";
import { Edit3, ChevronDown, ArrowUpCircle } from "lucide-react";

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
    <div className="w-[294px] h-[100px] bg-[#090909]/[0.55] backdrop-blur-xl rounded-[14px] border border-[#FFA845]/[0.10] p-4 flex flex-col justify-center gap-2">
      <div className="flex items-center justify-between">
        {isEditing ? (
          <input
            autoFocus
            className="w-full bg-[#090909]/[0.55] border border-[#FFA845]/[0.10] rounded px-2 py-1 text-sm text-gray-200 outline-none"
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
            <span className="text-sm font-medium text-gray-200">
              {username}
            </span>
            <div className="flex gap-2 items-center">

              <Edit3
                size={14}
                className="text-gray-400 cursor-pointer hover:text-white"
                onClick={() => onEditToggle(true)}
              />
            </div>
          </>
        )}
      </div>
      <div className="flex items-center justify-between bg-[#090909]/[0.55] backdrop-blur-md rounded-lg px-3 py-2 border border-white/5 cursor-pointer hover:bg-white/5">
        <span className="text-xs text-gray-300">
          {currentVersion || "Not installed"}
        </span>
        <ChevronDown size={14} className="text-gray-400" />
      </div>
    </div>
  );
};
