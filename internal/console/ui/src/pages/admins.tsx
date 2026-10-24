// The administrators of the server, for themselves: who they are, a new
// one with the link where they choose their password, a new link, and
// disabling. They hold every Cahut of the server, unlike the administrators
// of a Cahut, whom the screen of its people makes.
import { Plus } from "lucide-preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { Activation } from "../api.gen";
import { ConfirmButton, CopyField, Failure, initials, Notice, PageHead } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";
import { href, Link } from "../router";

const statuses: Record<string, { word: string; badge: string }> = {
	active: { word: "Active", badge: "c-badge--success" },
	pending: { word: "No password yet", badge: "c-badge--warning" },
	disabled: { word: "Disabled", badge: "c-badge--neutral" },
};

export function Admins() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.server("list_admins", {}));
	const [handle, setHandle] = useState("");
	const [name, setName] = useState("");
	const [failure, setFailure] = useState("");
	const [link, setLink] = useState<{ name: string; activation: Activation } | null>(null);
	if (error) return <Failure message={error} />;
	if (!data) return null;

	const act = async (run: () => Promise<{ name: string; activation: Activation } | null>) => {
		setFailure("");
		try {
			const given = await run();
			if (given) setLink(given);
			reload();
		} catch (err) {
			setFailure(explain(err));
		}
	};
	const create = (e: Event) => {
		e.preventDefault();
		act(async () => {
			const a = await api.server("create_admin", { handle, name });
			setHandle("");
			setName("");
			return { name: a.name, activation: a.activation };
		});
	};
	return (
		<>
			<PageHead label="This server" title="Server administrators">
				They hold every Cahut of this server: they declare Cahuts, set their kind and limits, read the Logbook of the server, and manage
				the people and Rooms of every Cahut, at its console.
			</PageHead>
			<div class="section">
				<Notice>
					To make someone an administrator of this Cahut only, add them in <Link class="c-link" href="/people">People</Link> as an
					administrator.
				</Notice>
			</div>
			<section class="section c-card panel">
				<h2 class="c-h4">New server administrator</h2>
				<form class="form" onSubmit={create}>
					<div class="form__row">
						<label class="c-field">
							<span class="c-field__label">Name</span>
							<input class="c-input" required value={name} onInput={(e) => setName(e.currentTarget.value)} />
						</label>
						<label class="c-field">
							<span class="c-field__label">Handle</span>
							<input class="c-input c-mono" required autocapitalize="none" value={handle} onInput={(e) => setHandle(e.currentTarget.value)} />
						</label>
					</div>
					<div>
						<button class="c-btn c-btn--primary" type="submit">
							<Plus class="i" size={16} aria-hidden="true" />
							Create the administrator
						</button>
					</div>
				</form>
			</section>
			{failure && <Failure message={failure} />}
			{link && (
				<div class="result section" role="status">
					<p class="c-small">
						Send {link.name} this link, where they choose their password. It works once, for 7 days, and is shown only now.
					</p>
					<CopyField label="the activation link" value={`${location.origin}${href("/server/activate")}#${link.activation.code}`} />
				</div>
			)}
			<div class="table-wrap section">
				<table class="table">
					<thead>
						<tr>
							<th>Administrator</th>
							<th>State</th>
							<th />
						</tr>
					</thead>
					<tbody>
						{data.admins.map((a) => (
							<tr key={a.handle}>
								<td>
									<span class="person">
										<span class="c-avatar" aria-hidden="true">
											{initials(a.name)}
										</span>
										<span class="people__who">
											<span>{a.name}</span>
											<span class="c-mono c-muted people__handle">{a.handle}</span>
										</span>
									</span>
								</td>
								<td>
									<span class={`c-badge ${(statuses[a.status] ?? statuses.active).badge}`}>{(statuses[a.status] ?? statuses.active).word}</span>
								</td>
								<td class="table__actions">
									{a.handle !== api.session.handle && a.status !== "disabled" && (
										<ConfirmButton
											label="New link"
											question={`Give ${a.name} a new link? Their password works until they use it.`}
											confirm="Give a link"
											onConfirm={() =>
												act(async () => ({ name: a.name, activation: await api.server("create_activation_link", { admin: a.handle }) }))
											}
										/>
									)}
									{a.handle !== api.session.handle &&
										(a.status === "disabled" ? (
											<ConfirmButton
												label="Enable"
												question={`Enable ${a.name} again?`}
												confirm="Enable"
												onConfirm={() => act(async () => (await api.server("enable_admin", { admin: a.handle }), null))}
											/>
										) : (
											<ConfirmButton
												label="Disable"
												question={`Disable ${a.name}? Their sessions end, and they cannot sign in until enabled again.`}
												confirm="Disable"
												danger
												onConfirm={() => act(async () => (await api.server("disable_admin", { admin: a.handle }), null))}
											/>
										))}
								</td>
							</tr>
						))}
					</tbody>
				</table>
			</div>
		</>
	);
}
