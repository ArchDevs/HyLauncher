import type React from "react";
import { Gamepad2, Globe, Globe2, Home, Server } from "lucide-react";

import HomePage from "../pages/Home";
import ServersPage from "../pages/Servers";

import BackgroundImage from "../components/BackgroundImage";
import BackgroundServers from "../components/BackgroundServers";

export type PageConfig = {
  id: string;
  name: string;
  icon: React.ComponentType<{ size?: number | string }>; // ✅ фикс для Lucide
  component: React.ComponentType;
  background?: React.ComponentType;
};

export const pages: PageConfig[] = [
  {
    id: "home",
    name: "Home",
    icon: Gamepad2,
    component: HomePage,
    background: BackgroundImage,
  },
  {
    id: "servers",
    name: "Servers",
    icon: Globe,
    component: ServersPage,
    background: BackgroundServers,
  },
];

export const getDefaultPage = () => pages[0];

export const getPageById = (id: string) => pages.find((p) => p.id === id);
