// The installer of Cahut, as it is at the root of the repository, so that
// one command installs a server:
//
//   curl -fsSL https://cahut.cloud/install.sh | sudo sh -s -- --domain cahut.example.com
//
// It checks the signature of the release it downloads. See
// docs/adr/0024-signed-releases-and-updates.md.
import script from '@repo/install.sh?raw';

export function GET(): Response {
	return new Response(script, { headers: { 'Content-Type': 'text/plain; charset=utf-8' } });
}
