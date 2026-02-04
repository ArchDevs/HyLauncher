import type React from "react";
import { Boxes, Gamepad2, Globe, Globe2, Home, Server } from "lucide-react";

import HomePage from "../pages/Home";
import ServersPage from "../pages/Servers";
import ModsPage from "../pages/Mods";

import BackgroundImage from "../components/BackgroundImage";
import BackgroundServers from "../components/BackgroundServers";
import type { Translations } from "../i18n/types";

export type PageConfigBase = {
  id: string;
  nameKey: keyof Translations["pages"];
  icon: React.ComponentType<{ size?: number | string }>;
  component: React.ComponentType;
  background?: React.ComponentType;
};

export type PageConfig = PageConfigBase & {
  name: string;
};

// Base pages config - no translations here, logic separated
const basePages: PageConfigBase[] = [
  {
    id: "home",
    nameKey: "home",
    icon: Gamepad2,
    component: HomePage,
    background: BackgroundImage,
  },
  {
    id: "servers",
    nameKey: "servers",
    icon: Globe,
    component: ServersPage,
    background: BackgroundServers,
  },
  {
    id: "mods",
    nameKey: "mods",
    icon: Boxes,
    component: ModsPage,
    background: BackgroundServers,
  },
];

/**
 * Get pages with translated names
 * This function bridges the gap between config and translations
 */
export const getPages = (t: Translations): PageConfig[] => {
  return basePages.map((page) => ({
    ...page,
    name: t.pages[page.nameKey],
  }));
};

export const getDefaultPage = (t: Translations) => getPages(t)[0];

export const getPageById = (id: string, t: Translations) =>
  getPages(t).find((p) => p.id === id);
