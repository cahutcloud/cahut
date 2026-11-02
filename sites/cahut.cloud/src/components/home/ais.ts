// The AI clients that publish to a Cahut, in the order of the dictionaries, as
// the help of their vendors says (docs.cahut.cloud/connect/ai-clients). Each
// shows the mark of its vendor, from its official files: see assets/ais.
import claude from '../../assets/ais/claude-spark.svg?raw';
import mistral from '../../assets/ais/mistral-icon.svg?raw';
import openai from '../../assets/ais/openai-blossom.svg?raw';

export const aiClients = [
	{ id: 'claude', free: true, mark: claude },
	{ id: 'mistral', free: true, mark: mistral },
	{ id: 'chatgpt', free: false, mark: openai },
] as const;

/** Which plans work, and how to connect each AI client. */
export const plansURL = 'https://docs.cahut.cloud/connect/ai-clients/#which-plans-work';
