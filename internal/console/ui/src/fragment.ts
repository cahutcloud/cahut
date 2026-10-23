// The fragment of the address, read once when the console loads: the codes
// of activation links and of the link to the first start come there, and the
// server never sees them. The console then takes it out of the address, so
// that it stays out of the browser's history.
export const fragment = location.hash.slice(1);
if (fragment) {
	history.replaceState(null, "", location.pathname + location.search);
}
