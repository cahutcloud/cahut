// Where to report a vulnerability, as RFC 9116 asks, with the same two
// channels as SECURITY.md. It expires 180 days after the build of the
// site, which every change to main runs again, so that it never names a
// channel that has stopped being read.
const lifetime = 180 * 24 * 60 * 60 * 1000;

export function GET(): Response {
	const expires = new Date(Date.now() + lifetime);
	expires.setUTCHours(0, 0, 0, 0);
	const text = [
		'Contact: mailto:security@cahut.cloud',
		'Contact: https://github.com/cahutcloud/cahut/security/advisories/new',
		`Expires: ${expires.toISOString()}`,
		'Preferred-Languages: en, fr',
		'Canonical: https://cahut.cloud/.well-known/security.txt',
		'Policy: https://cahut.cloud/project/security/',
		'',
	].join('\n');
	return new Response(text, { headers: { 'Content-Type': 'text/plain; charset=utf-8' } });
}
