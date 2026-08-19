let enabled = false;

async function request(path: string, init?: RequestInit): Promise<Response> {
  const response = await fetch(path, init);
  if (!response.ok) throw new Error(`${init?.method || "GET"} ${path}: ${response.status} ${await response.text()}`);
  return response;
}

const noteURL = (name: string) => `/api/notes?name=${encodeURIComponent(name)}`;

export async function initServerStorage(): Promise<boolean> {
  const response = await fetch("/api/notes");
  if (response.status !== 404 && !response.ok) throw new Error(`GET /api/notes: ${response.status} ${await response.text()}`);
  enabled = response.ok;
  return enabled;
}

export function isServerStorage(): boolean { return enabled; }
export async function serverLoadNoteNames(): Promise<string[]> { return request("/api/notes").then((r) => r.json()); }
export async function serverLoadNote(name: string): Promise<string> { return request(noteURL(name)).then((r) => r.text()); }
export async function serverSaveNote(name: string, content: string): Promise<void> { await request(noteURL(name), { method: "PUT", body: content }); }
export async function serverDeleteNote(name: string): Promise<void> { await request(noteURL(name), { method: "DELETE" }); }
export async function serverLoadMetadata(): Promise<string> { return request("/api/notes/metadata").then((r) => r.text()); }
export async function serverSaveMetadata(content: string): Promise<void> { await request("/api/notes/metadata", { method: "PUT", body: content }); }
