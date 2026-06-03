import type { Metadata } from "next";
import { JoinClient } from "./client";

export const metadata: Metadata = {
  title: "Join NUST Devs",
  description: "Request to add your GitHub profile to the NUST developer community.",
};

export default function JoinPage() {
  return <JoinClient />;
}
