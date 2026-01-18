import React, { useState } from 'react';
import { Edit3, ChevronDown, ArrowUpCircle, Plus, Trash2, Check, X } from 'lucide-react';
import { config } from '../../wailsjs/go/models';

interface ProfileProps {
  currentProfile: config.Profile | null;
  profiles: config.Profile[];
  currentVersion: string;
  updateAvailable: boolean;
  onUpdate: () => void;
  onProfileChange: (id: string) => void;
  onProfileAdd: (name: string) => void;
  onProfileDelete: (id: string) => void;
  onProfileUpdate: (id: string, name: string) => void;
}

export const ProfileSection: React.FC<ProfileProps> = ({
  currentProfile, profiles, currentVersion, updateAvailable, onUpdate,
  onProfileChange, onProfileAdd, onProfileDelete, onProfileUpdate
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [isAdding, setIsAdding] = useState(false);
  const [newName, setNewName] = useState("");
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState("");

  const handleAdd = () => {
    if (newName.trim()) {
      onProfileAdd(newName.trim());
      setNewName("");
      setIsAdding(false);
    }
  };

  const handleUpdate = (id: string) => {
    if (editName.trim()) {
      onProfileUpdate(id, editName.trim());
      setEditingId(null);
      setEditName("");
    }
  };

  return (
    <div className="relative w-[294px] flex flex-col gap-2">
      {/* Current Profile Card */}
      <div
        onClick={() => setIsOpen(!isOpen)}
        className="h-[100px] bg-[#090909]/[0.55] backdrop-blur-xl rounded-[14px] border border-[#FFA845]/[0.10] p-4 flex flex-col justify-center gap-2 cursor-pointer hover:bg-white/5 transition-colors group"
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2 overflow-hidden">
            <span className="text-sm font-medium text-gray-200 truncate max-w-[180px]">
              {currentProfile?.name || "Loading..."}
            </span>
            <div className="flex gap-2 items-center flex-shrink-0">
              {updateAvailable && (
                <button
                  onClick={(e) => { e.stopPropagation(); onUpdate(); }}
                  className="text-[#FFA845] hover:scale-110 transition-transform"
                >
                  <ArrowUpCircle size={16} />
                </button>
              )}
            </div>
          </div>
          <div className="flex items-center gap-2">
            <ChevronDown
              size={18}
              className={`text-gray-400 group-hover:text-white transition-transform ${isOpen ? 'rotate-180' : ''}`}
            />
          </div>
        </div>
        <div className="flex items-center justify-between bg-[#090909]/[0.55] backdrop-blur-md rounded-lg px-3 py-2 border border-white/5">
          <span className="text-xs text-gray-300">{currentVersion || "Not installed"}</span>
        </div>
      </div>


      {/* Dropdown Menu */}
      {isOpen && (
        <div className="absolute top-[110px] left-0 w-full bg-[#0d0d0d] border border-white/10 rounded-[14px] shadow-2xl z-50 overflow-hidden backdrop-blur-2xl">
          <div className="max-h-[240px] overflow-y-auto py-2">
            {profiles.map((p) => (
              <div
                key={p.id}
                className="group flex items-center justify-between px-4 py-2 hover:bg-white/5 transition-colors"
                onClick={() => !editingId && onProfileChange(p.id)}
              >
                {editingId === p.id ? (
                  <div className="flex items-center gap-2 w-full" onClick={(e) => e.stopPropagation()}>
                    <input
                      autoFocus
                      className="flex-1 bg-black/40 border border-[#FFA845]/30 rounded px-2 py-1 text-xs text-white outline-none"
                      value={editName}
                      onChange={(e) => setEditName(e.target.value)}
                      onKeyDown={(e) => e.key === 'Enter' && handleUpdate(p.id)}
                    />
                    <Check size={14} className="text-green-500 cursor-pointer" onClick={() => handleUpdate(p.id)} />
                    <X size={14} className="text-red-500 cursor-pointer" onClick={() => setEditingId(null)} />
                  </div>
                ) : (
                  <>
                    <span className={`text-xs ${p.id === currentProfile?.id ? 'text-[#FFA845] font-bold' : 'text-gray-300'}`}>
                      {p.name}
                    </span>
                    <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                      <Edit3
                        size={12}
                        className="text-gray-400 hover:text-white cursor-pointer"
                        onClick={(e) => { e.stopPropagation(); setEditingId(p.id); setEditName(p.name); }}
                      />
                      {profiles.length > 1 && (
                        <Trash2
                          size={12}
                          className="text-gray-400 hover:text-red-500 cursor-pointer"
                          onClick={(e) => { e.stopPropagation(); onProfileDelete(p.id); }}
                        />
                      )}
                    </div>
                  </>
                )}
              </div>
            ))}
          </div>

          <div className="border-t border-white/5 p-2">
            {isAdding ? (
              <div className="flex items-center gap-2">
                <input
                  autoFocus
                  placeholder="New profile name"
                  className="flex-1 bg-black/40 border border-[#FFA845]/30 rounded px-2 py-1 text-xs text-white outline-none"
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
                />
                <Check size={14} className="text-green-500 cursor-pointer" onClick={handleAdd} />
                <X size={14} className="text-red-500 cursor-pointer" onClick={() => setIsAdding(false)} />
              </div>
            ) : (
              <button
                onClick={() => setIsAdding(true)}
                className="w-full flex items-center justify-center gap-2 py-2 text-xs text-gray-400 hover:text-[#FFA845] transition-colors"
              >
                <Plus size={14} />
                <span>Add New Profile</span>
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
