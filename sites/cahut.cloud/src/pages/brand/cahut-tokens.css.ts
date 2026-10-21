// The CSS tokens of the brand, as they are in the guidelines. Imported with
// ?url, Vite would minify them and drop their comments.
import tokens from '@charte/tokens/cahut-tokens.css?raw';

export function GET(): Response {
	return new Response(tokens, { headers: { 'Content-Type': 'text/css; charset=utf-8' } });
}
