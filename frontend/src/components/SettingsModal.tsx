import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { Settings, X, Save, HardDrive, Monitor, Cpu, Folder, Loader2 } from 'lucide-react';
import { GetSettings, SaveSettings, GetGameVersions } from '../../wailsjs/go/app/App';
import { config } from '../../wailsjs/go/models';

interface SettingsModalProps {
    onClose: () => void;
}

export const SettingsModal: React.FC<SettingsModalProps> = ({ onClose }) => {
    const [activeTab, setActiveTab] = useState<'game' | 'java' | 'video'>('game');
    const [settings, setSettings] = useState<config.GameSettings | null>(null);
    const [availableVersions, setAvailableVersions] = useState<number[]>([]);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        loadSettings();
    }, []);

    useEffect(() => {
        if (settings?.channel) {
            loadVersions(settings.channel);
        }
    }, [settings?.channel]);

    const loadVersions = async (channel: string) => {
        try {
            const versions = await GetGameVersions(channel);
            setAvailableVersions(versions || []);
        } catch (err) {
            console.error("Failed to load versions:", err);
        }
    };

    const loadSettings = async () => {
        try {
            const data = await GetSettings();
            // Ensure we have a valid object instance
            setSettings(data);
        } catch (err) {
            console.error("Failed to load settings:", err);
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        if (!settings) return;
        setSaving(true);
        try {
            await SaveSettings(settings);
            onClose();
        } catch (err) {
            console.error("Failed to save settings:", err);
            setSaving(false);
        }
    };

    const updateSetting = (key: keyof config.GameSettings, value: any) => {
        if (!settings) return;
        const newSettings = new config.GameSettings(settings);
        // @ts-ignore
        newSettings[key] = value;
        setSettings(newSettings);
    };

    // Render loading state if needed, or just show the modal with a spinner overlay
    if (loading && !settings) {
        // You might want a better loading state
    }

    return (
        <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4"
            onClick={onClose}
        >
            <motion.div
                initial={{ scale: 0.9, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                exit={{ scale: 0.9, opacity: 0 }}
                onClick={(e) => e.stopPropagation()}
                className="w-full max-w-3xl bg-[#090909]/95 backdrop-blur-xl rounded-2xl border border-[#FFA845]/20 overflow-hidden shadow-2xl h-[600px] flex flex-col"
            >
                {/* Header */}
                <div className="p-6 border-b border-white/10 bg-gradient-to-r from-[#FFA845]/10 to-transparent flex justify-between items-center">
                    <div className="flex items-center gap-3">
                        <div className="p-2 rounded-lg bg-[#FFA845]/20">
                            <Settings size={24} className="text-[#FFA845]" />
                        </div>
                        <div>
                            <h3 className="text-lg font-bold text-white">Launcher Settings</h3>
                            <p className="text-xs text-gray-400">Configure your game experience</p>
                        </div>
                    </div>
                    <button onClick={onClose} className="p-2 hover:bg-white/10 rounded-lg transition-colors text-gray-400 hover:text-white">
                        <X size={20} />
                    </button>
                </div>

                {loading ? (
                    <div className="flex-1 flex items-center justify-center flex-col gap-4">
                        <Loader2 size={40} className="animate-spin text-[#FFA845]" />
                        <p className="text-gray-400 text-sm">Loading configuration...</p>
                    </div>
                ) : settings ? (
                    <div className="flex flex-1 overflow-hidden">
                        {/* Sidebar */}
                        <div className="w-48 bg-black/20 border-r border-white/5 p-4 flex flex-col gap-2">
                            <TabButton
                                active={activeTab === 'game'}
                                onClick={() => setActiveTab('game')}
                                icon={<HardDrive size={18} />}
                                label="Game"
                            />
                            <TabButton
                                active={activeTab === 'video'}
                                onClick={() => setActiveTab('video')}
                                icon={<Monitor size={18} />}
                                label="Video"
                            />
                            <TabButton
                                active={activeTab === 'java'}
                                onClick={() => setActiveTab('java')}
                                icon={<Cpu size={18} />}
                                label="Java"
                            />
                        </div>

                        {/* Content */}
                        <div className="flex-1 p-8 overflow-y-auto">
                            {activeTab === 'game' && (
                                <div className="space-y-6">
                                    <Section title="Game Directory" description="Location where the game files are stored">
                                        <div className="flex gap-2">
                                            <input
                                                type="text"
                                                value={settings.gameDir || 'Default'}
                                                disabled
                                                className="flex-1 bg-black/40 border border-white/10 rounded-lg px-4 py-2 text-sm text-gray-400 cursor-not-allowed"
                                            />
                                            <button className="px-4 py-2 bg-white/5 hover:bg-white/10 border border-white/10 rounded-lg text-gray-300 transition-colors">
                                                <Folder size={18} />
                                            </button>
                                        </div>
                                    </Section>

                                    <Section title="Update Settings" description="Configure game version and update channel">
                                        <div className="space-y-4">
                                            <div>
                                                <label className="text-xs text-gray-500 mb-1 block">Update Channel</label>
                                                <select
                                                    value={settings.channel || 'release'}
                                                    onChange={(e) => updateSetting('channel', e.target.value)}
                                                    className="w-full bg-black/40 border border-white/10 rounded-lg px-4 py-2 text-sm text-white focus:border-[#FFA845]/50 focus:outline-none transition-colors appearance-none"
                                                >
                                                    <option value="release">Release</option>
                                                    <option value="pre-release">Pre-Release</option>
                                                </select>
                                            </div>
                                            <div>
                                                <label className="text-xs text-gray-500 mb-1 block">Target Version</label>
                                                <select
                                                    value={settings.gameVersion}
                                                    onChange={(e) => updateSetting('gameVersion', parseInt(e.target.value))}
                                                    className="w-full bg-black/40 border border-white/10 rounded-lg px-4 py-2 text-sm text-white focus:border-[#FFA845]/50 focus:outline-none transition-colors appearance-none"
                                                >
                                                    <option value={0}>Latest (Auto-Update)</option>
                                                    {availableVersions.map(v => (
                                                        <option key={v} value={v}>Version {v}</option>
                                                    ))}
                                                </select>
                                            </div>
                                        </div>
                                    </Section>

                                    <div className="flex items-center justify-between bg-white/5 p-4 rounded-lg border border-white/5">
                                        <div>
                                            <h4 className="text-sm font-medium text-white">Online Fix</h4>
                                            <p className="text-xs text-gray-500">Installs Online-Fix in your game </p>
                                        </div>
                                        <button
                                            onClick={() => updateSetting('onlineFix', !settings.onlineFix)}
                                            className={`w-12 h-6 rounded-full transition-colors relative ${settings.onlineFix ? 'bg-[#FFA845]' : 'bg-gray-700'}`}
                                        >
                                            <div className={`absolute top-1 w-4 h-4 rounded-full bg-white transition-all ${settings.onlineFix ? 'left-7' : 'left-1'}`} />
                                        </button>
                                    </div>
                                </div>
                            )}

                            {activeTab === 'video' && (
                                <div className="space-y-6">
                                    <Section title="Resolution" description="Set the game window dimensions">
                                        <div className="grid grid-cols-2 gap-4">
                                            <div>
                                                <label className="text-xs text-gray-500 mb-1 block">Width</label>
                                                <input
                                                    type="number"
                                                    value={settings.width}
                                                    onChange={(e) => updateSetting('width', parseInt(e.target.value))}
                                                    className="w-full bg-black/40 border border-white/10 rounded-lg px-4 py-2 text-sm text-white focus:border-[#FFA845]/50 focus:outline-none transition-colors"
                                                />
                                            </div>
                                            <div>
                                                <label className="text-xs text-gray-500 mb-1 block">Height</label>
                                                <input
                                                    type="number"
                                                    value={settings.height}
                                                    onChange={(e) => updateSetting('height', parseInt(e.target.value))}
                                                    className="w-full bg-black/40 border border-white/10 rounded-lg px-4 py-2 text-sm text-white focus:border-[#FFA845]/50 focus:outline-none transition-colors"
                                                />
                                            </div>
                                        </div>
                                    </Section>

                                    <div className="flex items-center justify-between bg-white/5 p-4 rounded-lg border border-white/5">
                                        <div>
                                            <h4 className="text-sm font-medium text-white">Fullscreen</h4>
                                            <p className="text-xs text-gray-500">Launch the game in fullscreen mode</p>
                                        </div>
                                        <button
                                            onClick={() => updateSetting('fullscreen', !settings.fullscreen)}
                                            className={`w-12 h-6 rounded-full transition-colors relative ${settings.fullscreen ? 'bg-[#FFA845]' : 'bg-gray-700'}`}
                                        >
                                            <div className={`absolute top-1 w-4 h-4 rounded-full bg-white transition-all ${settings.fullscreen ? 'left-7' : 'left-1'}`} />
                                        </button>
                                    </div>
                                </div>
                            )}

                            {activeTab === 'java' && (
                                <div className="space-y-6">
                                    <Section title="Memory Allocation" description={`Allocate RAM for the game (${settings.minMemory}GB - ${settings.maxMemory}GB)`}>
                                        <div className="space-y-4">
                                            <div>
                                                <div className="flex justify-between text-xs text-gray-400 mb-2">
                                                    <span>Min: {settings.minMemory}GB</span>
                                                </div>
                                                <input
                                                    type="range"
                                                    min="1"
                                                    max="16"
                                                    value={settings.minMemory}
                                                    onChange={(e) => updateSetting('minMemory', parseInt(e.target.value))}
                                                    className="w-full h-2 bg-white/10 rounded-lg appearance-none cursor-pointer accent-[#FFA845]"
                                                />
                                            </div>
                                            <div>
                                                <div className="flex justify-between text-xs text-gray-400 mb-2">
                                                    <span>Max: {settings.maxMemory}GB</span>
                                                </div>
                                                <input
                                                    type="range"
                                                    min="1"
                                                    max="32"
                                                    step="1"
                                                    value={settings.maxMemory}
                                                    onChange={(e) => updateSetting('maxMemory', parseInt(e.target.value))}
                                                    className="w-full h-2 bg-white/10 rounded-lg appearance-none cursor-pointer accent-[#FFA845]"
                                                />
                                            </div>
                                        </div>
                                    </Section>

                                    <Section title="JVM Arguments" description="Advanced Java Virtual Machine arguments">
                                        <textarea
                                            value={settings.javaArgs}
                                            onChange={(e) => updateSetting('javaArgs', e.target.value)}
                                            className="w-full h-32 bg-black/40 border border-white/10 rounded-lg p-4 text-xs font-mono text-gray-300 focus:border-[#FFA845]/50 focus:outline-none transition-colors resize-none"
                                        />
                                    </Section>
                                </div>
                            )}
                        </div>
                    </div>
                ) : (
                    <div className="flex-1 flex items-center justify-center text-red-400">
                        Failed to load settings.
                    </div>
                )}

                {/* Footer */}
                <div className="p-6 border-t border-white/10 bg-black/20 flex justify-end gap-3">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 hover:bg-white/5 text-gray-400 hover:text-white rounded-lg transition-colors text-sm"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSave}
                        disabled={loading || saving || !settings}
                        className="px-6 py-2 bg-[#FFA845] hover:bg-[#ffb460] text-black font-bold rounded-lg transition-colors text-sm flex items-center gap-2 disabled:opacity-50"
                    >
                        {saving ? <Loader2 size={16} className="animate-spin" /> : <Save size={16} />}
                        {saving ? 'Saving...' : 'Save Changes'}
                    </button>
                </div>
            </motion.div>
        </motion.div>
    );
};

const TabButton = ({ active, onClick, icon, label }: { active: boolean, onClick: () => void, icon: any, label: string }) => (
    <button
        onClick={onClick}
        className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg transition-all text-sm font-medium ${active
            ? 'bg-[#FFA845]/20 text-[#FFA845] border border-[#FFA845]/20'
            : 'text-gray-400 hover:bg-white/5 hover:text-white border border-transparent'
            }`}
    >
        {icon}
        {label}
    </button>
);

const Section = ({ title, description, children }: { title: string, description: string, children: React.ReactNode }) => (
    <div className="bg-white/5 border border-white/5 rounded-xl p-5">
        <div className="mb-4">
            <h4 className="text-sm font-bold text-white mb-1">{title}</h4>
            <p className="text-xs text-gray-500">{description}</p>
        </div>
        {children}
    </div>
);
