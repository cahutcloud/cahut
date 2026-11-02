// The public key that the official MCP Registry reads here, alone and as
// plain text, to verify that the servers it lists under cloud.cahut, My Cahut
// and Cahut, come from the owner of cahut.cloud. The owner signs each login
// with the private key, which stays out of this repository. The key is not a
// secret: the Registry reads it at this address.
const record = 'v=MCPv1; k=ed25519; p=7IBQIrYN9Z5VfRD13wufsJmvT1NaZG9Z7IWpuh4kDF8=';

export function GET(): Response {
	return new Response(record, { headers: { 'Content-Type': 'text/plain; charset=utf-8' } });
}
