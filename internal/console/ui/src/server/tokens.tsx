// The service tokens of a Cahut, with which a program on the server, such
// as the site where people sign up, manages its people through the admin
// API. See docs/adr/0035-a-program-on-the-server-manages-people-with-a-service-token.md.
import { Plus } from "lucide-preact";
import { useState } from "preact/hooks";
import { type CahutCall, explain } from "../api";
import type { NewServiceToken } from "../api.gen";
import { ConfirmButton, CopyField, Failure, Notice } from "../components";
import { de, locale, words } from "../i18n";
import { useLoad } from "../load";

const t = words({
	en: {
		title: "Service tokens",
		lead: "A program on the server, such as the site where people sign up, manages the people of this Cahut with a service token, through the admin API. A token opens nothing else.",
		none: "No service token.",
		created: (date: string) => `Created on ${date}`,
		used: (date: string) => `, last used on ${date}`,
		revoked: (date: string) => `Revoked on ${date}`,
		revoke: "Revoke",
		revokeQuestion: "Revoke this token? The program that uses it loses its access at once.",
		name: "What uses it",
		create: "Create a token",
		shownOnce: (name: string) => `The token of ${name}. Copy it now: it is shown only once.`,
		token: "the token",
	},
	fr: {
		title: "Jetons de service",
		lead: "Un programme du serveur, comme le site où l'on s'inscrit, gère les personnes de cette Cahut avec un jeton de service, par l'API d'administration. Un jeton n'ouvre rien d'autre.",
		none: "Aucun jeton de service.",
		created: (date) => `Créé le ${date}`,
		used: (date) => `, utilisé pour la dernière fois le ${date}`,
		revoked: (date) => `Révoqué le ${date}`,
		revoke: "Révoquer",
		revokeQuestion: "Révoquer ce jeton ? Le programme qui s'en sert perd aussitôt son accès.",
		name: "Ce qui l'utilise",
		create: "Créer un jeton",
		shownOnce: (name) => `Le jeton ${de(name)}. Copiez-le maintenant : il n'est montré qu'une fois.`,
		token: "le jeton",
	},
});

function day(at: string): string {
	return new Date(at).toLocaleDateString(locale, { dateStyle: "medium" });
}

export function ServiceTokens({ call }: { call: CahutCall }) {
	const { data, error, reload } = useLoad(() => call("list_service_tokens", {}));
	const [name, setName] = useState("");
	const [made, setMade] = useState<NewServiceToken | null>(null);
	const [failure, setFailure] = useState("");
	const create = async (e: Event) => {
		e.preventDefault();
		setFailure("");
		try {
			setMade(await call("create_service_token", { name }));
			setName("");
			reload();
		} catch (err) {
			setFailure(explain(err));
		}
	};
	return (
		<section class="section c-card panel">
			<h2 class="c-h4">{t.title}</h2>
			<p class="c-small c-muted">{t.lead}</p>
			{error && <Failure message={error} />}
			{data && data.tokens.length === 0 && <Notice>{t.none}</Notice>}
			{data && data.tokens.length > 0 && (
				<ul class="tokens">
					{data.tokens.map((k) => (
						<li key={k.id} class="tokens__item">
							<span>
								<strong>{k.name}</strong>
								<span class="c-small c-muted">
									{" "}
									{k.revoked_at ? t.revoked(day(k.revoked_at)) : t.created(day(k.created_at)) + (k.last_used_at ? t.used(day(k.last_used_at)) : "")}
								</span>
							</span>
							{!k.revoked_at && (
								<ConfirmButton
									label={t.revoke}
									question={t.revokeQuestion}
									confirm={t.revoke}
									danger
									onConfirm={async () => {
										await call("revoke_service_token", { id: k.id });
										reload();
									}}
								/>
							)}
						</li>
					))}
				</ul>
			)}
			<form class="form" onSubmit={create}>
				<label class="c-field">
					<span class="c-field__label">{t.name}</span>
					<input class="c-input" required maxLength={100} value={name} onInput={(e) => setName(e.currentTarget.value)} />
				</label>
				{failure && <Failure message={failure} />}
				<div>
					<button class="c-btn c-btn--secondary" type="submit">
						<Plus class="i" size={16} aria-hidden="true" />
						{t.create}
					</button>
				</div>
			</form>
			{made && (
				<div class="result" role="status">
					<p class="c-small">{t.shownOnce(made.name)}</p>
					<CopyField label={t.token} value={made.token} />
				</div>
			)}
		</section>
	);
}
