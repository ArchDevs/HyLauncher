import { ComponentType } from "react";
import { LucideIcon } from "lucide-react";
import { Gamepad2, Globe } from "lucide-react";
import HomePage from "../pages/Home";
import ServersPage from "../pages/Servers";

export interface PageConfig {
  id: string;
  name: string;
  icon: LucideIcon;
  component: ComponentType;
  path: string; // For potential future URL routing
}

export const pages: PageConfig[] = [
  {
    id: "home",
    name: "Home",
    icon: Gamepad2,
    component: HomePage,
    path: "/",
  },
  {
    id: "servers",
    name: "Servers",
    icon: Globe,
    component: ServersPage,
    path: "/servers",
  },
];

// Helper function to get a page by ID
export const getPageById = (id: string): PageConfig | undefined => {
  return pages.find((page) => page.id === id);
};

// Helper function to get the default page
export const getDefaultPage = (): PageConfig => {
  return pages[0];
};

