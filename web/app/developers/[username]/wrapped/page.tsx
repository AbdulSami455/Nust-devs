import { WrappedClient } from "./client";

export default async function WrappedPage({
  params,
  searchParams,
}: {
  params: Promise<{ username: string }>;
  searchParams: Promise<{ year?: string }>;
}) {
  const { username } = await params;
  const { year } = await searchParams;
  return <WrappedClient username={username} year={year ? Number(year) : undefined} />;
}
