import React from "react";
import BannersHome from "../components/BannersHome";
import { ControlSection } from "../components/ControlSection";
import { ProfileSection } from "../components/ProfileCard";
import { UpdateOverlay } from "../components/UpdateOverlay";
import { DeleteConfirmationModal } from "../components/DeleteConfirmationModal";
import { ErrorModal } from "../components/ErrorModal";
import { useLauncher } from "../hooks/useLauncher";
import { OpenFolder, DeleteGame } from "../../wailsjs/go/app/App";

function HomePage() {
  const {
    username,
    currentVersion,
    launcherVersion,
    isEditingUsername,
    setIsEditingUsername,
    progress,
    status,
    isDownloading,
    downloadDetails,
    updateAsset,
    isUpdatingLauncher,
    updateStats,
    showDeleteModal,
    setShowDeleteModal,
    setShowDiagnostics,
    error,
    setError,
    handlePlay,
    handleUpdateLauncher,
    setNick,
  } = useLauncher();

  return (
    <>
      {isUpdatingLauncher && (
        <UpdateOverlay
          progress={progress}
          downloaded={updateStats.d}
          total={updateStats.t}
        />
      )}

      <main className="relative z-10 h-full p-10 flex flex-col justify-between pt-[60px]">
        <div className="flex justify-between items-start">
          <ProfileSection
            username={username}
            currentVersion={currentVersion}
            isEditing={isEditingUsername}
            onEditToggle={setIsEditingUsername}
            onUserChange={setNick}
          />

          <BannersHome />
        </div>

        <ControlSection
          onPlay={handlePlay}
          isDownloading={isDownloading}
          progress={progress}
          status={status}
          speed={downloadDetails.speed}
          downloaded={downloadDetails.downloaded}
          total={downloadDetails.total}
          currentFile={downloadDetails.currentFile}
          actions={{
            openFolder: OpenFolder,
            showDiagnostics: () => setShowDiagnostics(true),
            showDelete: () => setShowDeleteModal(true),
          }}
          updateAvailable={!!updateAsset}
          onUpdate={handleUpdateLauncher}
        />

        <div className="absolute right-[16px] bottom-[16px] text-[#FFFFFF]/[0.25] text-[14px] font-[Mazzard]">
          {launcherVersion}v
        </div>
      </main>

      {showDeleteModal && (
        <DeleteConfirmationModal
          onConfirm={() => {
            DeleteGame("default");
            setShowDeleteModal(false);
          }}
          onCancel={() => setShowDeleteModal(false)}
        />
      )}

      {error && <ErrorModal error={error} onClose={() => setError(null)} />}
    </>
  );
}

export default HomePage;
