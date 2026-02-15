import { useState } from "react";
import { AnimatePresence } from "framer-motion";
import Banner from "../components/Banner";
import { useTranslation } from "../i18n";
import { useLauncher, ServerWithFullUrls } from "../hooks/useLauncher";
import ServerModal from "../components/ServerModal";

function ServersPage() {
  const { t } = useTranslation();
  const { servers, isLoadingServers, handlePlay } = useLauncher();
  const [selectedServer, setSelectedServer] = useState<ServerWithFullUrls | null>(null);

  // Show all servers, fill remaining slots with placeholders
  const displayServers = servers;
  const placeholderCount = Math.max(0, 4 - displayServers.length);

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

      {/* Loading State */}
      {isLoadingServers && (
        <div className="absolute left-[88px] top-[100px] flex flex-wrap gap-x-[22px] gap-y-[22px]">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="w-[448px] h-[200px] rounded-[20px] bg-[#090909]/55 backdrop-blur-[12px] animate-pulse border border-[#7C7C7C]/10" />
          ))}
        </div>
      )}

      {/* Servers Grid */}
      <div className="absolute left-[88px] top-[100px] flex flex-wrap gap-x-[22px] gap-y-[22px]">
        {/* Large Server Banners */}
        {displayServers.map((server) => (
          <Banner
            key={server.id}
            variant="large"
            backgroundImage={server.banner_url}
            iconImage={server.logo_url}
            title={server.name}
            description={server.description}
            onClick={() => setSelectedServer(server)}
          />
        ))}

        {/* Placeholder slots for advertising */}
        {Array.from({ length: placeholderCount }).map((_, index) => (
          <Banner
            key={`placeholder-${index}`}
            variant="small"
            text={t.banners?.advertising || "По поводу рекламы пишите нашему боту @hylauncher_bot"}
          />
        ))}
      </div>

      {/* Server Detail Modal */}
      <AnimatePresence>
        {selectedServer && (
          <ServerModal
            server={selectedServer}
            isOpen={!!selectedServer}
            onClose={() => setSelectedServer(null)}
            onPlay={handlePlay}
          />
        )}
      </AnimatePresence>
    </div>
  );
}

export default ServersPage;
