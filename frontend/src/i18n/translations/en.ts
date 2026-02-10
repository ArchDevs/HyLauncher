import type { Translations } from "../types";

export const en: Translations = {
  common: {
    play: "PLAY",
    install: "INSTALL...",
    ready: "Ready",
    cancel: "Cancel",
    close: "Close",
    delete: "Delete",
    confirm: "Confirm",
    update: "Update",
    updateAvailable: "Update available",
    updating: "Updating",
    error: "Error",
    copy: "Copy",
    copied: "Copied!",
  },
  pages: {
    home: "Home",
    servers: "Servers",
    mods: "Mods",
  },
  profile: {
    username: "Username",
    version: "Version",
    noVersion: "No",
    releaseType: {
      preRelease: "Pre-Release",
      release: "Release",
    },
    loading: "Loading",
  },
  control: {
    status: {
      readyToPlay: "Ready to play",
    },
    updateAvailable: "Update available",
  },
  modals: {
    delete: {
      title: "Are you sure?",
      message: "Do you really want to delete the game?",
      warning:
        "This action will delete all game files without the possibility of recovery!",
      confirmButton: "Delete all",
      cancelButton: "Cancel",
    },
    error: {
      title: "An error occurred",
      technicalDetails: "Technical details",
      stackTrace: "Stack trace",
      suggestion: "Please report this issue if it persists.",
      copyError: "Copy error",
      copied: "Copied!",
      suggestions: {
        network: "Check your internet connection and try again.",
        filesystem:
          "Make sure you have enough disk space and the launcher has proper permissions.",
        validation: "Please check your input and try again.",
        game: "Try restarting the launcher or reinstalling the game.",
        default: "Please report this issue if it persists.",
      },
    },
    update: {
      title: "UPDATING LAUNCHER",
      message:
        "Downloading the latest version. HyLauncher will restart automatically once finished.",
    },
    server: {
      copyIp: "Copy IP",
      copied: "Copied!",
      play: "Play",
    },
  },
  banners: {
    advertising: "Contact @hylauncher_bot for advertising",
    noServers: "No servers available",
    hynexus: {
      text: "HyNexus - this is Hytale as it should be. Economy, Clans, PVP, PVE, we're waiting for you!",
    },
    nctale: {
      text: "Join the community on Discord and Telegram — news, help and chat.",
    },
  },
  settings: {
    note: "Note:",
    translationNotice: "The app is not fully translated yet, so some content may remain in English for certain languages.",
  },
};
