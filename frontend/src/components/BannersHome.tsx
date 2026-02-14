import { useState } from "react";
import { AnimatePresence } from "framer-motion";
import { useTranslation } from "../i18n";
import { useLauncher, ServerWithFullUrls } from "../hooks/useLauncher";
import Banner from "./Banner";
import ServerModal from "./ServerModal";

interface BannersHomeProps {
  servers: ServerWithFullUrls[];
  isLoading: boolean;
  onPlay?: (serverIP: string) => void;
}

function BannersHome({ servers, isLoading, onPlay }: BannersHomeProps) {
  const { t } = useTranslation();
  const [selectedServer, setSelectedServer] = useState<ServerWithFullUrls | null>(null);

  // Show up to 5 banners
  const displayServers = servers.slice(0, 5);

  return (
    <div className="flex flex-col gap-[10px] max-h-[340px] overflow-y-auto scrollbar-hide">
      {/* Loading state - show placeholder banners */}
      {isLoading && (
        <>
          <div className="w-[400px] h-[80px] rounded-[20px] border border-[#FFA845]/10 bg-[#090909]/55 backdrop-blur-[12px] animate-pulse flex-shrink-0" />
          <div className="w-[400px] h-[80px] rounded-[20px] border border-[#FFA845]/10 bg-[#090909]/55 backdrop-blur-[12px] animate-pulse flex-shrink-0" />
        </>
      )}

      {/* Server banners - up to 5 */}
      {displayServers.map((server) => (
        <div key={server.id} className="flex-shrink-0">
          <Banner
            variant="compact"
            iconImage={server.logo_url}
            title={`${server.name} — ${server.ip}`}
            description={server.description}
            onClick={() => setSelectedServer(server)}
          />
        </div>
      ))}

      {/* Fallback if no servers */}
      {!isLoading && displayServers.length === 0 && (
        <div className="w-[400px] h-[80px] rounded-[20px] border border-[#FFA845]/10 bg-[#090909]/55 backdrop-blur-[12px] flex items-center justify-center flex-shrink-0">
          <span className="text-[14px] text-[#CCD9E0]/50 font-[Mazzard]">
            {t.banners?.noServers || "No servers available"}
          </span>
        </div>
      )}

      {/* Server Detail Modal */}
      <AnimatePresence>
        {selectedServer && (
          <ServerModal
            server={selectedServer}
            isOpen={!!selectedServer}
            onClose={() => setSelectedServer(null)}
            onPlay={onPlay}
          />
        )}
      </AnimatePresence>
    </div>
  );
}

export default BannersHome;
