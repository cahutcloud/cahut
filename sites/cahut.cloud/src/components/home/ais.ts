// The AI clients that publish to a Cahut, in the order of the dictionaries, as
// the help of their vendors says (docs.cahut.cloud/connect/ai-clients). They
// go by their names alone: their vendors give their logos to partners only.
// Each links to its film on YouTube (videos.ts), but ChatGPT, whose film waits.
export const aiClients = [
	{ id: 'claude', free: true, film: 'claude' },
	{ id: 'mistral', free: true, film: 'mistral' },
	{ id: 'chatgpt', free: false, film: null },
] as const;

/** Which plans work, and how to connect each AI client. */
export const plansURL = 'https://docs.cahut.cloud/connect/ai-clients/#which-plans-work';
