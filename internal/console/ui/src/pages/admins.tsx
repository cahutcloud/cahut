// The administrators of the server, for themselves: who they are, a new
// one with the link where they choose their password, and a new link.
import { Plus } from "lucide-preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { Activation } from "../api.gen";
import { ConfirmButton, CopyField, Failure, initials, PageHead } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";

export function Admins() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.server("list_admins", {}));
	const [handle, setHandle] = useState("");
	const [name, setName] = useState("");
	const [failure, setFailure] = useState("");
	const [link, setLink] = useState<{ name: string; activation: Activation } | null>(null);
	if (error) return <Failure message={error} />;
	if (!data) return null;

	const act = async (run: () => Promise<{ name: string; activation: Activation }>) => {
		setFailure("");
		try {
			setLink(await run());
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
			<PageHead label="This server" title="Administrators">
				They declare Cahuts, read the Logbook of the server, and manage the people and Rooms of every Cahut, at its console.
			</PageHead>
			<section class="section c-card panel">
				<h2 class="c-h4">New administrator</h2>
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
					<CopyField label="the activation link" value={`${location.origin}/server/activate#${link.activation.code}`} />
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
									<span class={`c-badge ${a.status === "active" ? "c-badge--success" : "c-badge--warning"}`}>
										{a.status === "active" ? "Active" : "No password yet"}
									</span>
								</td>
								<td class="table__actions">
									{a.handle !== api.session.handle && (
										<ConfirmButton
											label="New link"
											question={`Give ${a.name} a new link? Their password works until they use it.`}
											confirm="Give a link"
											onConfirm={() =>
												act(async () => ({ name: a.name, activation: await api.server("create_activation_link", { admin: a.handle }) }))
											}
										/>
									)}
								</td>
							</tr>
						))}
					</tbody>
				</table>
			</div>
		</>
	);
}
