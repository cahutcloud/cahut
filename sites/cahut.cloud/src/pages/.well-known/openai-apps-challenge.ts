// The token that OpenAI gives the owner of a domain to serve here, alone and
// as plain text, to verify that a plugin of its directory comes from them. It
// proves that the plugin of my.cahut.cloud, a name under cahut.cloud, which
// cannot serve the token itself, comes from the owner of cahut.cloud. The
// token is not a secret: OpenAI reads it at this address.
const token = 'SC80RZuy89vOo8As3xCq6unEwlFtAD-7wGdNxlahLZE';

export function GET(): Response {
	return new Response(token, { headers: { 'Content-Type': 'text/plain; charset=utf-8' } });
}
