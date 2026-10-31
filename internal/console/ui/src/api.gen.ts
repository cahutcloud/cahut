// Generated from the Go types of the registries by make console-types: do not edit.
// See docs/adr/0014-console-http-api.md.

export interface About {
  domain: string;
  /** the address of the Cahut, under which its Rooms and Windows open */
  url?: string;
  /** the address of its MCP endpoint, where AI clients connect */
  mcp?: string;
  /** whether this Cahut sends e-mail, through the relay that its administrators set; without one, nothing that would send e-mail works, such as invitations */
  sends_mail: boolean;
  /** collaborative, where people share team Rooms, or individual, where each person works alone and does not see the others */
  kind: string;
  /** the limits that the administrators of the server set; a limit that is absent counts nothing. A member of an individual Cahut gets the limits per person only */
  limits: Limits;
  /** what the Cahut holds, as its limits count it; absent for a member of an individual Cahut */
  usage?: Usage;
  /** in an individual Cahut, what the person you act for holds, as the limits per person count it */
  yours?: PersonUsage;
  /** the links to the legal pages of the Cahut, which its host publishes elsewhere, and which the console and the pages for readers show at their foot; empty when it links none */
  legal: LegalLinks;
  /** since when and why the administrators of the server suspended the Cahut; absent while it is active */
  suspension?: Suspension;
  /** the address that receives each report about the Windows of the Cahut, beside its administrators, which the administrators of the server set; absent when it has none, and for anyone but its administrators */
  moderation_email?: string;
}

export interface AccessInput {
  /** the Window's id or address, or its Room and its name as room/name */
  window: string;
  /** author, room, internal, key or public */
  access: string;
  /** when the access ends, such as 2026-12-31T18:00:00Z; after it, the Window opens to its Room only. It does not end when absent */
  ends_at?: string;
  /** for key: the Key to set, of 10 to 256 characters. When empty, the Window keeps its Key, or Cahut draws one */
  key?: string;
  /** for key: draw a new Key; the old one no longer opens the Window */
  new_key?: boolean;
}

export interface AccessSet {
  /** the Window's id, which never changes: name the Window by it to find it whatever its name */
  id: string;
  /** the Window's name, last part of its address */
  name: string;
  title: string;
  /** what the Window is, in a sentence or two; absent when it has none */
  description?: string;
  /** its tags, in order, which help find it among many */
  tags: string[];
  kind: string;
  /** where the Window opens */
  url: string;
  /** the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts */
  permanent_url: string;
  /** the name of its Room, the part of its address before its name */
  room: string;
  /** the number of the version that readers see */
  version: number;
  /** the number of the version that rollback_window goes back to; absent when there is none */
  previous_version?: number;
  /** total size of the files of the current version, in bytes */
  size: number;
  file_count: number;
  /** handle of the person it was first published for, empty once that person is deleted */
  created_by: string;
  created_at: string;
  updated_at: string;
  /** who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public */
  access: string;
  /** when the access ends, after which the Window opens to its Room only; absent when it does not end */
  access_ends_at?: string;
  /** when the Window went to the trash; absent when it is not in it */
  trashed_at?: string;
  /** what the pages of the current version do, as declared when it was published; absent for a Window published before declarations */
  declaration?: Declaration;
  /** when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn */
  withdrawn?: Withdrawal;
  /** for a search, a few words of its pages around the words found; absent when its pages hold none */
  excerpt?: string;
  /** the new Key, returned this once: give it to the person, who passes it on with the address */
  key?: string;
}

export interface Account {
  id: number;
  handle: string;
  name: string;
  /** admin or member of the Cahut */
  role: string;
  /** pending until the person chooses a password, then active, or disabled */
  status: string;
  /** name of the person's personal Room, the handle */
  personal_room: string;
  /** the e-mail address of the person, which only the person and the administrators of the Cahut read */
  email?: string;
  /** whether the e-mail notices that are neither about the account nor about a Knock are off */
  notices_off?: boolean;
  /** the e-mail address that waits for the person to confirm it, from the link sent to it */
  email_change?: EmailChange;
  /** the Rooms of the person, their personal Room first */
  rooms: Membership[];
}

export interface AccountSession {
  /** send it in the Authorization header: Bearer <token> */
  token: string;
  /** when the session ends at the latest; it ends sooner after 7 days without use */
  expires_at: string;
  user: User;
}

export interface ActivateInput {
  /** the code of the activation link */
  code: string;
  /** the new password, of 15 characters at least */
  password: string;
}

export interface Activation {
  /** handle of the person */
  user: string;
  /** the person opens the console of the Cahut at /activate#<code> to choose their password */
  code: string;
  expires_at: string;
  /** the link went to the e-mail address of the person too */
  emailed?: boolean;
}

export interface ActivationLinkInput {
  /** handle of the person */
  user: string;
  /** false keeps the link from the e-mail address of the person, when you give it to them yourself; true when absent */
  send_activation_email?: boolean;
}

export interface AddDomainInput {
  /** the domain of the new Cahut, such as cahut.example.com, or the public IPv4 address of the server, for a Cahut without a domain */
  domain: string;
  /** the first administrator of the Cahut, a person of it */
  admin?: FirstAdmin;
}

export interface Addressee {
  /** the recipient, as the invitation names them */
  to: string;
}

export interface Admin {
  handle: string;
  name: string;
  /** pending until the administrator chooses a password, then active; disabled while they may not sign in */
  status: string;
  created_at: string;
}

export interface AdminInput {
  /** handle of the administrator */
  admin: string;
}

export interface AdminList {
  admins: Admin[];
}

export interface Answer {
  knock: Knock;
  /** what the operation returned, when it ran */
  output?: unknown;
}

export interface AnswerInput {
  /** the number of the Knock */
  knock: number;
}

export interface ApproveInput {
  request: Request;
  /** read, or write; at most what the client asks for */
  scope: string;
  /** every Room of the person, including those they join later */
  all_rooms?: boolean;
  /** the Rooms granted, by name, when all_rooms is false */
  rooms?: string[];
}

export interface At {
  /** the Room of the Window */
  room: string;
  /** the name of the Window */
  window: string;
}

export interface Authorization {
  /** the name of the AI client, as it gives it */
  client: string;
  /** its web page, as it gives it */
  client_uri?: string;
  /** the host of its client metadata document, which vouches for it */
  client_host?: string;
  /** where the browser goes back to after consent */
  redirect_host: string;
  /** the most the client asks for: read, or write */
  scope: string;
  /** the Rooms the person may grant */
  rooms: RoomChoice[];
  /** when set, the request is invalid: send the browser to this address */
  redirect?: string;
  /** why the request is invalid, when it is */
  error?: string;
}

export interface ByAuthor {
  /** the Room of the Window */
  room: string;
  /** the name of the Window */
  window: string;
  /** the handle of its author */
  author: string;
}

export interface ChangeDataInput {
  /** the Window's id or address, or room/name */
  window: string;
  /** SQLite statements, separated by semicolons, that create, add, change or delete */
  sql: string;
  /** their named parameters */
  params?: Record<string, number[]>;
}

export interface ChangeEmailInput {
  /** your new e-mail address, which a link sent to it confirms; empty removes yours */
  email: string;
}

export interface ChangeHandleInput {
  /** your new handle, which names your personal Room too */
  handle: string;
}

export interface ChangeNameInput {
  /** your name, as the people of the Cahut see it */
  name: string;
}

export interface ChangePasswordInput {
  /** the current password */
  password: string;
  /** the new password, of 15 characters at least */
  new_password: string;
}

export interface ChangedData {
  statements: number;
  changes: number;
  removes?: string[];
  /** the copy made before, which rollback_window_data goes back to */
  copy?: string;
}

export interface Check {
  at: string;
  /** the newest published release */
  latest?: string;
  /** the page of that release, with its notes */
  url?: string;
  /** why the releases could not be read */
  error?: string;
}

export interface CheckHandleInput {
  handle: string;
}

export interface Column {
  name: string;
  type?: string;
  not_null?: boolean;
  primary_key?: boolean;
}

export interface ConfirmEmailInput {
  /** the code of the link sent to the new address */
  code: string;
}

export interface ConnectedAIs {
  connections: Connection[];
}

export interface Connection {
  id: number;
  /** the person the AI acts for */
  user: string;
  /** the name of the AI client */
  client: string;
  client_id: string;
  /** where its authorization went back to */
  redirect_host: string;
  /** read, or write */
  scope: string;
  /** every Room of the person, including those they join later */
  all_rooms: boolean;
  /** the Rooms granted, when all_rooms is false */
  rooms: string[];
  /** active, expired after 30 days without use, or ended */
  status: string;
  created_at: string;
  last_used_at?: string;
  ended_at?: string;
  end_reason?: string;
}

export interface CreateAdminInput {
  handle: string;
  name: string;
}

export interface CreateRoomInput {
  /** title of the Room, such as Marketing */
  title: string;
  /** what the Room is, in a sentence or two of plain text, 300 characters at most */
  description?: string;
  /** name of the Room, the first part of its address; from the title when empty */
  name?: string;
  /** handle of its first owner; you when empty. Only administrators of the Cahut name someone else */
  owner?: string;
}

export interface CreateServiceTokenInput {
  /** what uses the token, such as the site where people sign up */
  name: string;
}

export interface CreateSetupLinkInput {
  /** the Cahut, by its domain or its address */
  domain: string;
}

export interface CreateTokenInput {
  /** handle of the person the AI acts for; yourself when empty. Only administrators create tokens for others */
  user?: string;
  /** what the token is for, such as the AI that uses it */
  name: string;
}

export interface CreateUserInput {
  handle: string;
  name: string;
  admin?: boolean;
  /** the e-mail address of the person, where their activation link goes */
  email?: string;
  /** false keeps the activation link from the e-mail address of the person, when you give it to them yourself; true when absent */
  send_activation_email?: boolean;
}

export interface Data {
  /** the SQL that makes the database, in order: create tables, indexes, views and triggers, add columns, insert rows. Each runs once: keep those that ran as they are, add new ones at the end */
  migrations?: string[];
  /** the tables, and their columns, that readers SELECT themselves at .data/sql; other columns read as NULL. List nothing private */
  readable?: Record<string, string[]>;
  /** the statements that the pages call at .data/queries/<name>, named with lowercase letters, digits and _ */
  queries?: Record<string, Query>;
  /** when the queries that write stop */
  closes?: string;
}

export interface DataCopy {
  name: string;
  /** when it was made */
  at: string;
  size: number;
}

export interface DataInfo {
  /** its tables, with their columns and how many rows each holds */
  tables: Table[];
  views?: string[];
  indexes?: string[];
  triggers?: string[];
  /** how many migrations of the declaration ran */
  migrations: number;
  /** its size, in bytes */
  size: number;
  /** the copies that rollback_window_data goes back to, the latest first */
  copies: DataCopy[];
}

export interface Declaration {
  /** whether the pages run scripts, inline or in files */
  scripts: boolean;
  /** every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make */
  origins?: string[];
  /** the origins that the forms of the pages send to, if any */
  forms?: string[];
  /** a SQLite database of the Window on the server, for a poll, a form or a tracker, which its pages call at .data/ under its root */
  data?: Data;
}

export interface DeleteRoomInput {
  /** the name of the team Room */
  room: string;
}

export interface DeletedRoom {
  /** the name of the Room, now free */
  room: string;
  title: string;
  /** how many Windows were deleted with it, those in its trash included */
  windows: number;
}

export interface DeletedUser {
  /** the handle of the person, now free */
  user: string;
  name: string;
  /** the Rooms deleted with the person: their personal Room, and the team Rooms they were the only owner of */
  rooms: string[];
  /** how many Windows were deleted with these Rooms, those in their trash included */
  windows: number;
}

export interface Delivery {
  /** the relay of the Cahut; absent when there is none, and the Cahut then sends no e-mail */
  relay?: RelayInfo;
  /** whether the Cahut sends e-mail: it has a relay, or the server runs for development and writes its messages to disk */
  sends: boolean;
  /** when a message last went out */
  last_sent?: string;
  /** why the last message that failed did not go out */
  last_error?: string;
  /** when it failed */
  last_error_at?: string;
}

export interface DenyInput {
  request: Request;
}

export interface DescribeInput {
  request: Request;
}

export type DescribeSignInInput = Record<string, never>;

export interface DescriptionLength {
  /** how many characters it has */
  length: number;
  /** the most it may have */
  limit: number;
}

export interface DisconnectAIInput {
  /** the connection, from list_connected_ais */
  id: number;
}

export interface DismissInput {
  /** the number of the report, which list_reports gives */
  report: number;
  /** why the Window stays as it is, for the moderators, 500 characters at most */
  note?: string;
}

export interface DomainList {
  cahuts: Info[];
}

export interface EmailChange {
  email: string;
  /** when the link sent to it stops working */
  expires_at: string;
}

export interface EmailChanged {
  /** your e-mail address now */
  email?: string;
  /** the address that waits for you to confirm it, from the link sent to it */
  waiting?: EmailChange;
}

export interface Entry {
  id: number;
  at: string;
  operation: string;
  /** admin, person or ai */
  actor: string;
  /** the person, or the person an AI acts for */
  user_id?: number;
  user?: string;
  /** the personal token the AI presented */
  token_id?: number;
  /** the connection of the AI, when it connected over OAuth */
  connection_id?: number;
  /** the AI, as its token or its connection, and its software, name it */
  client?: string;
  /** the adapter the call came through: cli, api, mcp, oauth, or web for the page of a share link */
  via: string;
  /** what the operation acted on */
  target?: string;
  detail?: string;
  /** ok, knocked (waiting for a person's approval), refused or failed */
  outcome: string;
  error?: string;
  /** the Knock that the call made, answered or ran for */
  knock_id?: number;
  /** where a call that signs in came from: an IPv4 address, or the /64 of an IPv6 one */
  address?: string;
}

export interface File {
  /** relative path with / as separator, such as index.html or img/logo.svg */
  path: string;
  content: string;
  /** utf-8 (the default) or base64, for binary files */
  encoding?: string;
}

export interface FileInfo {
  path: string;
  size: number;
  /** media type */
  type: string;
  content?: string;
  /** utf-8 or base64 */
  encoding?: string;
}

export interface FirstAdmin {
  handle: string;
  name: string;
}

export type GetAccountInput = Record<string, never>;

export type GetCahutInput = Record<string, never>;

export type GetMailRelayInput = Record<string, never>;

export interface GivenName {
  /** the name */
  name: string;
}

export interface GivenTag {
  /** the tag, as it was given */
  tag: string;
}

export interface Handle {
  /** the handle, as it was given */
  handle: string;
}

export interface HandleCheck {
  handle: string;
  /** the handle can go to a new person */
  available: boolean;
  /** why it cannot: invalid, reserved or taken */
  reason?: string;
}

export interface Info {
  id: number;
  domain: string;
  created_at: string;
  /** since when and why the administrators of the server suspended this Cahut, which serves nothing meanwhile; absent while it is active */
  suspension?: Suspension;
  /** how many reports about the Windows of this Cahut wait for its moderators */
  reports?: number;
}

export interface Invitation {
  id: number;
  /** the handle of the person invited, or the address */
  to: string;
  /** the handle of the person who invited */
  by?: string;
  /** the AI that sent the invitation for them */
  ai?: string;
  /** whether it carries a personal link, which opens the Window for this recipient alone */
  link: boolean;
  created_at: string;
  /** when its personal link stops working */
  expires_at?: string;
  /** when it was revoked, or ended with the Key or the access it stood for */
  ended_at?: string;
}

export interface InvitationCodeInput {
  /** the code of the personal link of the invitation */
  code: string;
}

export interface InvitationList {
  /** the latest first */
  invitations: Invitation[];
}

export interface InvitationOpening {
  code: string;
  /** the Window, as room/name */
  window: string;
}

export interface InvitationPage {
  /** the Window, as room/name */
  window: string;
  title: string;
  /** the name of the person who invites */
  by: string;
  /** the AI that sent the invitation for them */
  ai?: string;
  note?: string;
  /** the note in paragraphs and lists, as the page shows it */
  note_blocks?: NoteBlock[];
  expires_at: string;
}

export interface InviteInput {
  /** the Window's id or address, or its Room and its name as room/name */
  window: string;
  /** who to invite, 20 at most: handles of people of the Cahut, who sign in to read it, or e-mail addresses of anyone else, for a Window with a Key or public */
  to: string[];
  /** a few words for them, 500 characters at most, in plain text with **bold**, *italics*, and lists whose lines start with - or 1. */
  note?: string;
}

export interface Invited {
  window: string;
  invitations: Invitation[];
  /** people of the Cahut who have no e-mail address, to whom nothing went */
  unreachable?: string[];
}

export interface Invitee {
  /** the handle of the person */
  handle: string;
}

export interface Knock {
  id: number;
  /** the operation that the AI called */
  operation: string;
  /** the arguments of the call */
  input: Record<string, unknown>;
  /** what the call acts on */
  target?: string;
  detail?: string;
  /** the name of the Room the call concerns, if any */
  room?: string;
  /** handle of the person the AI acts for, who answers the Knock */
  user: string;
  /** the AI that called, as its token or its connection, and its software, name it */
  client: string;
  /** waiting, done (approved and run), failed (approved, and the operation failed), denied, expired or canceled (the AI lost access) */
  state: string;
  created_at: string;
  /** when the Knock expires, if no one answers it */
  expires_at: string;
  answered_at?: string;
  /** why the operation failed, once approved */
  error?: string;
  /** where the person approves or denies it, in the console */
  url?: string;
}

export interface KnockGetInput {
  /** the number of the Knock */
  knock: number;
}

export interface KnockList {
  knocks: Knock[];
  /** how many Knocks wait, in all */
  waiting: number;
}

export interface KnockListInput {
  /** only the Knocks that wait */
  waiting?: boolean;
  /** how many Knocks, 50 by default, at most 500 */
  limit?: number;
}

export interface LegalLink {
  link: string;
}

export interface LegalLinks {
  /** the address of the legal notice: who publishes the service and who hosts it */
  legal_notice?: string;
  /** the address of the terms of use */
  terms?: string;
  /** the address of the privacy policy */
  privacy?: string;
  /** the address of the page where people get help */
  support?: string;
  /** the addresses of the same pages in French, where they differ: the console links them when it speaks French. Each needs the address of its page for every other language */
  fr?: LegalPages;
}

export interface LegalPages {
  /** the address of the legal notice: who publishes the service and who hosts it */
  legal_notice?: string;
  /** the address of the terms of use */
  terms?: string;
  /** the address of the privacy policy */
  privacy?: string;
  /** the address of the page where people get help */
  support?: string;
}

export interface Length {
  /** how many characters it has */
  length: number;
  /** the fewest, or the most, it may have */
  limit: number;
}

export interface Limit {
  /** the most characters it may have */
  limit: number;
}

export interface Limits {
  /** how many people the Cahut may have, active or waiting for their activation, administrators included; disabled people do not count */
  people?: number;
  /** how many of them may be administrators of the Cahut */
  admins?: number;
  /** how many team Rooms the Cahut may have, or in an individual Cahut how many Rooms beside the personal ones; personal Rooms come with their person and do not count */
  rooms?: number;
  /** how many Windows out of the trash the Cahut may hold */
  windows?: number;
  /** how many Windows out of the trash each Room may hold, personal Rooms included */
  windows_per_room?: number;
  /** how many bytes the Cahut may keep: every version of every Window, those in the trash included */
  disk?: number;
  /** in an individual Cahut, how many Rooms each person may have beside their personal Room */
  rooms_per_person?: number;
  /** in an individual Cahut, how many Windows out of the trash each person may hold, over their Rooms */
  windows_per_person?: number;
  /** in an individual Cahut, how many bytes the Windows of each person may keep, every version included */
  disk_per_person?: number;
}

export interface LinkEnd {
  /** the end given, in RFC 3339 */
  ends_at: string;
  /** the latest end that a link made now may have, in RFC 3339 */
  latest: string;
}

export interface LinkInFrenchOnly {
  link: string;
  page: string;
}

export interface LinkLabel {
  /** the most characters of a label */
  limit: number;
}

export interface LinkOpenings {
  given: number;
  limit: number;
}

export interface LinkRef {
  /** the id of the link */
  link: number;
}

export interface LinksMade {
  /** how many share links were made in the last 24 hours */
  made: number;
  /** the person who makes them, or whom the AI acts for */
  handle: string;
  /** the most links of one person in 24 hours */
  limit: number;
}

export interface LinksOn {
  /** the Room of the Window */
  room: string;
  /** the name of the Window */
  window: string;
  /** the most links that stand on a Window */
  limit: number;
}

export type ListAdminsInput = Record<string, never>;

export interface ListConnectedAIsInput {
  /** only the AIs of this person; for a person who is not an administrator, their own */
  user?: string;
}

export type ListDomainsInput = Record<string, never>;

export interface ListInvitationsInput {
  /** the Window's id or address, or its Room and its name as room/name */
  window: string;
}

export interface ListLogbookInput {
  /** how many entries, 50 by default, at most 1000 */
  limit?: number;
}

export interface ListPublicInput {
  /** how many Windows, 100 by default, at most 500 */
  limit?: number;
}

export interface ListReportsInput {
  /** only the reports that wait for a decision */
  waiting?: boolean;
  /** how many reports, 100 by default, at most 500 */
  limit?: number;
}

export interface ListRoomsInput {
  /** every Room of the Cahut, for an administrator; your Rooms when false */
  all?: boolean;
}

export type ListServiceTokensInput = Record<string, never>;

export interface ListTokensInput {
  /** only the tokens of this person; for a person who is not an administrator, their own tokens */
  user?: string;
}

export type ListUsersInput = Record<string, never>;

export interface LogbookPage {
  entries: Entry[];
}

export interface MailNotices {
  /** whether the notices that are neither about your account nor about a Knock are off */
  off: boolean;
}

export interface MemberInput {
  /** the Room's name */
  room: string;
  /** handle of the person */
  user: string;
  /** owner or member; member when empty */
  role?: string;
}

export interface Membership {
  /** the Room's name, first part of its address */
  room: string;
  title: string;
  /** personal or team */
  kind: string;
  /** owner or member of the Room */
  role: string;
}

export interface Moderated {
  window: Reported;
  /** when and why it is withdrawn; absent once it is reinstated */
  withdrawn?: Withdrawal;
  /** the reports that waited about it, which the withdrawal decided */
  decided?: number[];
}

export interface ModerationAddressSet {
  /** the address of moderation of the Cahut; absent when it has none */
  email?: string;
}

export interface MovedRoom {
  /** the name of the Room now */
  room: string;
  /** the name it would take */
  name: string;
}

export interface NameInRoom {
  /** the name */
  name: string;
  /** the name of the Room */
  room: string;
}

export interface NewAdmin {
  handle: string;
  name: string;
  /** pending until the administrator chooses a password, then active; disabled while they may not sign in */
  status: string;
  created_at: string;
  /** the link where the administrator chooses their password */
  activation: Activation;
}

export interface NewCahut {
  id: number;
  domain: string;
  created_at: string;
  /** since when and why the administrators of the server suspended this Cahut, which serves nothing meanwhile; absent while it is active */
  suspension?: Suspension;
  /** how many reports about the Windows of this Cahut wait for its moderators */
  reports?: number;
  /** the first administrator of the Cahut, with their activation link */
  admin?: NewUser;
}

export interface NewName {
  /** the name of the Room */
  name: string;
}

export interface NewServiceToken {
  id: number;
  name: string;
  token: string;
}

export interface NewToken {
  id: number;
  user: string;
  name: string;
  token: string;
}

export interface NewUser {
  id: number;
  handle: string;
  name: string;
  /** admin or member of the Cahut */
  role: string;
  /** pending until the person chooses a password, then active, or disabled */
  status: string;
  /** name of the person's personal Room, the handle */
  personal_room: string;
  /** the e-mail address of the person, which only the person and the administrators of the Cahut read */
  email?: string;
  /** the link where the person chooses their password; an AI never gets it */
  activation?: Activation;
}

export interface NoOwner {
  /** the name of the Room */
  room: string;
}

export interface NoteBlock {
  /** bullets or numbers for a list, empty for a paragraph */
  list?: string;
  /** the number of the first item of a numbered list */
  start?: number;
  /** the lines of a paragraph, or the items of a list */
  lines: NoteRun[][];
}

export interface NoteLimit {
  /** the most characters of a note */
  limit: number;
}

export interface NoteRun {
  text: string;
  bold?: boolean;
  italic?: boolean;
}

export interface OpenShareLinkInput {
  /** the token of the share link, the last part of its address */
  token: string;
}

export interface Outsider {
  /** the handle of the person */
  handle: string;
  /** the Room of the Window */
  room: string;
  /** the name of the Window */
  window: string;
}

export interface PageData {
  columns: string[];
  rows: number[][][];
  cut?: boolean;
  changes?: number;
  id?: number;
}

export interface PageDataInput {
  /** room/name of a public Window, as its address has it */
  window?: string;
  /** the capability of the frame that opened the Window, as its address has it */
  frame?: string;
  /** for select_window_data: one SELECT over what the declaration makes readable */
  sql?: string;
  /** for call_window_query: the name of a query of the declaration */
  query?: string;
  /** the named parameters of the statement; :reader is given by Cahut */
  params?: Record<string, number[]>;
  no_write?: boolean;
}

export type PasswordChanged = Record<string, never>;

export interface PasswordLinkRequested {
  message: string;
}

export interface Person {
  handle: string;
  name: string;
  /** owner or member of the Room */
  role: string;
  /** how many Windows of the Room they keep for themselves, which open to the Room if they leave it; given to them, to the owners of the Room and to the administrators */
  kept?: number;
}

export interface PersonUsage {
  /** Rooms beside the personal one */
  rooms: number;
  /** Windows out of the trash, over their Rooms */
  windows: number;
  /** bytes of every version of the Windows of their Rooms, those in the trash included */
  disk: number;
}

export interface Preview {
  /** the Window, as room/name */
  window: string;
  /** the path of the frame that shows the Window, on the domain of the Cahut, such as /-/frame/<capability>/ */
  frame: string;
  /** when the frame stops showing the Window, if the session has not ended before */
  expires_at: string;
}

export interface PreviewInput {
  /** the Windows to preview, by their address or as room/name, at most 100 */
  windows: string[];
}

export interface PreviewReportedInput {
  /** the number of a report that waits, which list_reports gives */
  report: number;
}

export interface Previews {
  /** the previews of the Windows asked for that are in your Rooms and out of the trash, in the same order */
  previews: Preview[];
}

export interface PublicWindow {
  /** the Window's id, which never changes */
  id: string;
  room: string;
  name: string;
  title: string;
  /** author, room, internal, key or public */
  access: string;
  /** the Window is in the trash */
  trashed?: boolean;
  /** the moderators withdrew the Window, which serves nothing to its readers until they reinstate it */
  withdrawn?: boolean;
  /** the handle of the person it was first published for; empty once they are deleted */
  by?: string;
  /** when its access was last set, or when it was published */
  public_at: string;
}

export interface PublicWindows {
  /** the public Windows of the Cahut, out of the trash and not withdrawn, the latest made public first */
  windows: PublicWindow[];
}

export interface PublishInput {
  /** title of the Window, also used in its address */
  title: string;
  /** what the Window is, in a sentence or two of plain text, 300 characters at most: readers see it on the page of its Room, and AIs read it to find the Window */
  description?: string;
  /** 10 at most, such as q3 or sales, which help find it among many: reuse those that list_windows gives */
  tags?: string[];
  /** one of document, deck, dashboard or site */
  kind: string;
  /** name of the Room to publish in; the personal Room of the person you act for when empty */
  room?: string;
  /** its name in its address, 1 to 63 lowercase letters, digits and hyphens; from the title when empty */
  name?: string;
  /** room, the members of its Room (the default), or author, the person you act for alone: a draft, or a Window that one reader opens once with create_share_link and max_openings 1 */
  access?: string;
  /** the files of the Window; index.html at the root is the page that opens. Link them with relative paths, such as css/deck.css */
  files: File[];
  /** what the pages do: Cahut lets them load and call the origins listed here, and nothing else */
  declaration: Declaration;
}

export interface QueriedData {
  columns: string[];
  rows: number[][][];
  cut?: boolean;
  note: string;
}

export interface Query {
  /** one statement that reads, adds, changes or deletes rows; parameters such as :name come from the page, :reader from Cahut */
  sql: string;
  /** anyone, people signed in to the Cahut, or members of its Room */
  who: string;
  /** reads, adds, changes or deletes, which Cahut sets */
  kind?: string;
}

export interface QueryDataInput {
  /** the Window's id or address, or room/name */
  window: string;
  /** one SQLite SELECT over its whole database */
  sql: string;
  /** its named parameters, such as {"city": "Lyon"} for :city */
  params?: Record<string, number[]>;
}

export interface ReaderCode {
  /** send the browser to <domain>/-/session?code=<code>, on the port of the Windows */
  code: string;
  expires_at: string;
}

export interface ReasonLimit {
  /** the most characters of a reason */
  limit: number;
}

export interface Redirect {
  /** send the browser to this address */
  redirect: string;
}

export interface RefInput {
  /** the Window's id or address, or its Room and its name as room/name */
  window: string;
}

export interface ReinstateInput {
  /** the Window, by its id or as room/name */
  window: string;
}

export interface RelayInfo {
  host: string;
  port: number;
  /** tls or starttls */
  security: string;
  username?: string;
  /** whether a password is kept; it is never given back */
  password: boolean;
  /** the address that messages come from */
  from: string;
  /** the name that messages come from; the domain of the Cahut when absent */
  name?: string;
  set_at: string;
}

export interface RemoveMemberInput {
  /** the Room's name */
  room: string;
  /** handle of the person; yourself to leave the Room */
  user: string;
}

export interface RenameAdminInput {
  /** handle of the administrator */
  admin: string;
  /** their new name */
  name: string;
}

export interface RenameDomainInput {
  /** the domain or the address of the Cahut now */
  domain: string;
  /** its new domain, such as cahut.example.com, which must point to this server, or the public IPv4 address of the server */
  to: string;
}

export interface RenameInput {
  /** the Window's id or address, or its Room and its name as room/name */
  window: string;
  /** the new name, 1 to 63 lowercase letters, digits and hyphens */
  name: string;
}

export interface Report {
  id: number;
  window: Reported;
  /** what is wrong: phishing, malware, illegal, rights or other */
  category: string;
  explanation: string;
  /** the name of whoever reported, until the decision */
  name?: string;
  /** the address of whoever reported, until the decision */
  email?: string;
  created_at: string;
  /** dismissed or withdrawn; absent while it waits */
  decision?: string;
  decided_at?: string;
  /** the handle of the moderator who decided */
  decided_by?: string;
  /** an administrator of the server decided */
  by_server?: boolean;
  /** what the moderators noted when they left the Window as it is */
  note?: string;
}

export interface ReportInput {
  /** the address of the Window, as its reader meets it: its address, a file under it, its permanent address or a share link, such as https://cahut.example.com/room/window/ */
  window: string;
  /** what is wrong: phishing, malware, illegal, rights or other */
  category: string;
  /** what is wrong with the Window, from 10 to 2000 characters */
  explanation: string;
  /** the name of whoever reports, for the moderators */
  name?: string;
  /** where the receipt and the decision go */
  email?: string;
  /** whoever reports says that they do it in good faith, and that what they say is true */
  good_faith: boolean;
}

export interface ReportReceived {
  /** the number of the report */
  report: number;
}

export interface ReportRef {
  /** the number of the report */
  report: number;
}

export interface Reported {
  /** the Window's id, which never changes */
  id: string;
  room: string;
  name: string;
  title: string;
  /** author, room, internal, key or public */
  access: string;
  /** the Window is in the trash */
  trashed?: boolean;
  /** the moderators withdrew the Window, which serves nothing to its readers until they reinstate it */
  withdrawn?: boolean;
}

export interface ReportedPreview {
  /** the Window, as room/name */
  window: string;
  /** the path of the frame that shows the Window, on the domain of the Cahut, such as /-/frame/<capability>/ */
  frame: string;
  /** when the frame stops showing the Window, if the session or the report have not ended before */
  expires_at: string;
}

export interface Reports {
  /** how many reports wait for a decision, in all */
  waiting: number;
  /** the reports that wait, the oldest first, then the latest decided */
  reports: Report[];
}

export interface Request {
  response_type?: string;
  client_id: string;
  redirect_uri?: string;
  scope?: string;
  state?: string;
  code_challenge?: string;
  code_challenge_method?: string;
  resource?: string;
}

export interface RequestPasswordLinkInput {
  /** the handle of the account, or its e-mail address */
  who: string;
}

export interface RequestUpdateInput {
  /** a published release newer than the version running, such as 0.4.1; empty to check for new releases only */
  version?: string;
}

export interface RequestedUpdate {
  /** check, or the version to install */
  requested: string;
}

export type ResumeCahutInput = Record<string, never>;

export interface RetitleInput {
  /** the Window's id or address, or its Room and its name as room/name */
  window: string;
  /** the new title; the title stays when empty. The address does not change */
  title?: string;
  /** the new description, a sentence or two of plain text, 300 characters at most; empty to remove it, and it stays when absent */
  description?: string;
  /** the new tags, in place of the old ones, 10 at most; an empty list removes them, and they stay when absent */
  tags?: string[];
}

export interface RevokeInvitationInput {
  /** the id of the invitation, which list_invitations gives */
  invitation: number;
}

export interface RevokeServiceTokenInput {
  id: number;
}

export interface RevokeShareLinkInput {
  /** the id of the share link, which get_window gives */
  link: number;
}

export interface RevokeTokenInput {
  id: number;
}

export interface RollbackDataInput {
  /** the Window's id or address, or room/name */
  window: string;
  /** the name of a copy, as get_window lists them; the latest when empty */
  copy?: string;
}

export interface RolledBackData {
  /** the copy the database went back to */
  restored: DataCopy;
  /** the copy of the database as it was, which another rollback_window_data goes back to */
  kept: string;
}

export interface Room {
  /** the Room's name, first part of its address */
  name: string;
  title: string;
  /** what the Room is, in a sentence or two; absent when it has none */
  description?: string;
  /** personal or team */
  kind: string;
  /** your role in the Room, owner or member; empty when you are not in it */
  role?: string;
  /** how many people are in the Room */
  members: number;
  /** how many Windows of the Room you see, out of the trash: all but those that others keep for themselves */
  windows: number;
  /** the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them */
  home?: string;
  created_at: string;
}

export interface RoomChoice {
  name: string;
  title: string;
  /** personal or team */
  kind: string;
}

export interface RoomDetail {
  /** the Room's name, first part of its address */
  name: string;
  title: string;
  /** what the Room is, in a sentence or two; absent when it has none */
  description?: string;
  /** personal or team */
  kind: string;
  /** your role in the Room, owner or member; empty when you are not in it */
  role?: string;
  /** how many people are in the Room */
  members: number;
  /** how many Windows of the Room you see, out of the trash: all but those that others keep for themselves */
  windows: number;
  /** the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them */
  home?: string;
  created_at: string;
  /** the people in the Room, owners first */
  people: Person[];
}

export interface RoomGetInput {
  /** the Room's name, first part of its address */
  room: string;
}

export interface RoomList {
  rooms: Room[];
}

export interface Run {
  id: string;
  from: string;
  to: string;
  /** command line, timer or console */
  by: string;
  started: string;
  ended: string;
  /** updated, rolled_back, refused or failed */
  result: string;
  /** why it did not update */
  reason?: string;
}

export interface SendTestMailInput {
  /** on the command line, or for an administrator of the server, the address to send the test to; a person's test goes to their own address */
  to?: string;
}

export interface Sent {
  /** how many invitations went out in the last 24 hours */
  sent: number;
  /** the person who invites, or whom the AI acts for */
  handle: string;
  /** the most invitations of one person in 24 hours */
  limit: number;
}

export interface ServerSession {
  /** send it in the Authorization header: Bearer <token> */
  token: string;
  /** the domain where the session holds */
  domain: string;
  /** when the session ends at the latest; it ends sooner after 7 days without use */
  expires_at: string;
  admin: Admin;
}

export interface ServiceTokenInfo {
  id: number;
  name: string;
  created_at: string;
  last_used_at?: string;
  revoked_at?: string;
}

export interface ServiceTokenList {
  tokens: ServiceTokenInfo[];
}

export interface SetAIReachInput {
  /** the connection, from list_connected_ais; an AI leaves it out, to set its own */
  id?: number;
  /** read, or write */
  scope: string;
  /** every Room of the person, including those they join later */
  all_rooms?: boolean;
  /** the names of the Rooms the AI reaches, when all_rooms is false */
  rooms?: string[];
}

export interface SetCahutSettingsInput {
  /** collaborative or individual; the kind stays when empty. A Cahut becomes individual only when no Room has more than one member and no Window opens to the whole Cahut */
  kind?: string;
  /** every limit of the Cahut: a limit left out is lifted */
  limits: Limits;
}

export interface SetHomeInput {
  /** the Room's name */
  room: string;
  /** name of a Window of the Room, shown at the Room's address; empty to go back to the page that lists the Room's Windows */
  window?: string;
}

export interface SetLegalLinksInput {
  /** the address of the legal notice: who publishes the service and who hosts it */
  legal_notice?: string;
  /** the address of the terms of use */
  terms?: string;
  /** the address of the privacy policy */
  privacy?: string;
  /** the address of the page where people get help */
  support?: string;
  /** the addresses of the same pages in French, where they differ: the console links them when it speaks French. Each needs the address of its page for every other language */
  fr?: LegalPages;
}

export interface SetMailNoticesInput {
  /** true turns off the notices that are neither about your account nor about a Knock, such as being added to a Room; false turns them on again */
  off: boolean;
}

export interface SetMailRelayInput {
  /** the name or the address of the SMTP relay, such as smtp.example.com, without a scheme or a port; empty removes the relay */
  host: string;
  /** its port: 465 for TLS from the start, 587 for STARTTLS; 465 with tls and 587 otherwise when left out */
  port?: number;
  /** tls, from the start, or starttls; tls on port 465 and starttls otherwise when left out */
  security?: string;
  /** the user name at the relay; leave it out when the relay takes no authentication */
  username?: string;
  /** the password at the relay; when left out, the password kept stays, as long as the host and the user name do not change */
  password?: string;
  /** the address that messages come from, one that the relay accepts, such as cahut@example.com */
  from?: string;
  /** the name that messages come from; the domain of the Cahut when left out */
  name?: string;
}

export interface SetModerationAddressInput {
  /** the address that receives each report about the Windows of this Cahut, such as abuse@host.example; empty removes it */
  email?: string;
}

export interface SetUp {
  /** the session of the first person, at the console of the Cahut */
  session: AccountSession;
  /** whether they are also the administrator of the server, with the same handle and password */
  server_admin: boolean;
}

export interface SetUpInput {
  /** the code of the link to the first start */
  code: string;
  handle: string;
  name: string;
  /** of 15 characters at least */
  password: string;
}

export interface SetUserEmailInput {
  /** handle of the person */
  user: string;
  /** their e-mail address; empty removes it */
  email: string;
}

export interface SetUserRoleInput {
  /** handle of the person */
  user: string;
  /** admin or member */
  role: string;
}

export interface Settings {
  /** collaborative, where people share team Rooms, or individual, where each person works alone and does not see the others */
  kind: string;
  /** the limits that the administrators of the server set; a limit that is absent counts nothing. A member of an individual Cahut gets the limits per person only */
  limits: Limits;
  /** what the Cahut holds, as its limits count it; absent for a member of an individual Cahut */
  usage?: Usage;
  /** in an individual Cahut, what the person you act for holds, as the limits per person count it */
  yours?: PersonUsage;
  /** the links to the legal pages of the Cahut, which its host publishes elsewhere, and which the console and the pages for readers show at their foot; empty when it links none */
  legal: LegalLinks;
  /** since when and why the administrators of the server suspended the Cahut; absent while it is active */
  suspension?: Suspension;
  /** the address that receives each report about the Windows of the Cahut, beside its administrators, which the administrators of the server set; absent when it has none, and for anyone but its administrators */
  moderation_email?: string;
}

export interface Setup {
  domain: string;
  expires_at: string;
  /** whether the first person also becomes the administrator of the server, which has none */
  server_admin: boolean;
}

export interface SetupInput {
  /** the code of the link to the first start */
  code: string;
}

export interface SetupLink {
  domain: string;
  /** the first person opens the console of the Cahut at /setup#<code> */
  code: string;
  expires_at: string;
  /** whether the first person, today, would also become the administrator of the server, which has none */
  server_admin: boolean;
}

export interface ShareLink {
  id: number;
  label?: string;
  /** the handle of the person who made it */
  by?: string;
  /** the AI that made it for them */
  ai?: string;
  created_at: string;
  /** when it stops opening the Window, to those who opened it too */
  expires_at: string;
  /** how many readers may open it; absent when any number may */
  max_openings?: number;
  /** how many readers opened it */
  openings: number;
  /** when it was last opened */
  opened_at?: string;
  /** when it was taken back */
  revoked_at?: string;
  /** whether it opens the Window to someone new: neither ended nor taken back, with openings left */
  stands: boolean;
}

export interface ShareLinkInput {
  /** the Window's id or address, or its Room and its name as room/name */
  window: string;
  /** when the link stops opening the Window, within 30 days, such as 2026-11-05T18:00:00Z; 7 days from now when absent */
  ends_at?: string;
  /** how many readers may open it, from 1 to 1000; any number when absent */
  max_openings?: number;
  /** a few words that tell the links of the Window apart, such as For Acme, 100 characters at most, which only the members of its Room see */
  label?: string;
}

export interface ShareLinkMade {
  id: number;
  label?: string;
  /** the handle of the person who made it */
  by?: string;
  /** the AI that made it for them */
  ai?: string;
  created_at: string;
  /** when it stops opening the Window, to those who opened it too */
  expires_at: string;
  /** how many readers may open it; absent when any number may */
  max_openings?: number;
  /** how many readers opened it */
  openings: number;
  /** when it was last opened */
  opened_at?: string;
  /** when it was taken back */
  revoked_at?: string;
  /** whether it opens the Window to someone new: neither ended nor taken back, with openings left */
  stands: boolean;
  /** the Window, as room/name */
  window: string;
  /** the share link, returned this once: give it to the person, who passes it on. Whoever holds it opens the Window */
  link: string;
}

export interface ShareLinkOpening {
  /** the Window, as room/name */
  window: string;
  /** when the pass ends: with the link, or 4 hours after the opening of a link that opens once */
  expires_at: string;
}

export interface SignInInput {
  handle: string;
  password: string;
}

export interface SignInPage {
  /** whether this Cahut sends e-mail, and so a link to a person who forgot their password */
  sends_mail: boolean;
  /** the links to the legal pages of the Cahut, which the pages before sign-in show at their foot */
  legal: LegalLinks;
  /** whether the administrators of the server suspended this Cahut: nobody but them signs in until they resume it */
  suspended?: boolean;
}

export type SignInToWindowsInput = Record<string, never>;

export type SignOutInput = Record<string, never>;

export type SignedOut = Record<string, never>;

export interface Started {
  /** the domain of the Cahut */
  domain: string;
}

export interface SuspendCahutInput {
  /** why the Cahut is suspended, in 500 characters at most, such as an unpaid invoice or a decision of moderation; its administrators read it, and readers, people and AIs never do */
  reason: string;
}

export interface Suspension {
  /** when the administrators of the server suspended the Cahut */
  since: string;
  /** why the administrators of the server suspended the Cahut; its administrators read it, and readers, people and AIs never do */
  reason: string;
}

export interface Table {
  name: string;
  columns: Column[];
  rows: number;
}

export interface TagCount {
  /** how many tags were given */
  count: number;
  /** the most a Window carries */
  limit: number;
}

export interface TakenEmail {
  email: string;
}

export interface TestSent {
  to: string;
}

export interface TokenInfo {
  id: number;
  user: string;
  name: string;
  created_at: string;
  last_used_at?: string;
  revoked_at?: string;
}

export interface TokenList {
  tokens: TokenInfo[];
}

export interface UpdateInput {
  /** the Window's id or address, or its Room and its name as room/name */
  window: string;
  /** a new title; the title stays when empty. The address does not change */
  title?: string;
  /** a new description, a sentence or two of plain text; empty to remove it, and it stays when absent */
  description?: string;
  /** every file of the new version, as for publish_window; files left out are not in it */
  files: File[];
  /** what the pages do: Cahut lets them load and call the origins listed here, and nothing else */
  declaration: Declaration;
}

export interface UpdateRoomInput {
  /** the Room's name */
  room: string;
  /** the new title; the title stays when empty */
  title?: string;
  /** the new description, a sentence or two of plain text, 300 characters at most; empty to remove it, and it stays when absent */
  description?: string;
}

export interface Updates {
  /** the version of Cahut running */
  running: string;
  /** whether the server updates itself, as install.sh sets it up */
  managed: boolean;
  /** what the daily run does: patch (installs the patch releases of the minor version running), all, check or off */
  policy?: string;
  check?: Check;
  last?: Run;
  /** what an administrator asked for, check or a version, which the updater has not taken yet */
  requested?: string;
}

export type UpdatesInput = Record<string, never>;

export interface Usage {
  /** people who are not disabled, administrators included */
  people: number;
  /** administrators who are not disabled */
  admins: number;
  /** team Rooms, or in an individual Cahut the Rooms beside the personal ones */
  rooms: number;
  /** Windows out of the trash */
  windows: number;
  /** bytes of every version of every Window, those in the trash included */
  disk: number;
}

export interface User {
  id: number;
  handle: string;
  name: string;
  /** admin or member of the Cahut */
  role: string;
  /** pending until the person chooses a password, then active, or disabled */
  status: string;
  /** name of the person's personal Room, the handle */
  personal_room: string;
  /** the e-mail address of the person, which only the person and the administrators of the Cahut read */
  email?: string;
}

export interface UserInput {
  /** handle of the person */
  user: string;
}

export interface UserList {
  users: User[];
}

export interface Version {
  number: number;
  /** whether readers see this version */
  current: boolean;
  /** total size of its files, in bytes */
  size: number;
  file_count: number;
  /** handle of the person it was published for, empty once that person is deleted */
  published_by: string;
  published_at: string;
}

export interface Wait {
  /** how long to wait, in minutes */
  minutes: number;
}

export interface Window {
  /** the Window's id, which never changes: name the Window by it to find it whatever its name */
  id: string;
  /** the Window's name, last part of its address */
  name: string;
  title: string;
  /** what the Window is, in a sentence or two; absent when it has none */
  description?: string;
  /** its tags, in order, which help find it among many */
  tags: string[];
  kind: string;
  /** where the Window opens */
  url: string;
  /** the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts */
  permanent_url: string;
  /** the name of its Room, the part of its address before its name */
  room: string;
  /** the number of the version that readers see */
  version: number;
  /** the number of the version that rollback_window goes back to; absent when there is none */
  previous_version?: number;
  /** total size of the files of the current version, in bytes */
  size: number;
  file_count: number;
  /** handle of the person it was first published for, empty once that person is deleted */
  created_by: string;
  created_at: string;
  updated_at: string;
  /** who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public */
  access: string;
  /** when the access ends, after which the Window opens to its Room only; absent when it does not end */
  access_ends_at?: string;
  /** when the Window went to the trash; absent when it is not in it */
  trashed_at?: string;
  /** what the pages of the current version do, as declared when it was published; absent for a Window published before declarations */
  declaration?: Declaration;
  /** when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn */
  withdrawn?: Withdrawal;
  /** for a search, a few words of its pages around the words found; absent when its pages hold none */
  excerpt?: string;
}

export interface WindowDetail {
  /** the Window's id, which never changes: name the Window by it to find it whatever its name */
  id: string;
  /** the Window's name, last part of its address */
  name: string;
  title: string;
  /** what the Window is, in a sentence or two; absent when it has none */
  description?: string;
  /** its tags, in order, which help find it among many */
  tags: string[];
  kind: string;
  /** where the Window opens */
  url: string;
  /** the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts */
  permanent_url: string;
  /** the name of its Room, the part of its address before its name */
  room: string;
  /** the number of the version that readers see */
  version: number;
  /** the number of the version that rollback_window goes back to; absent when there is none */
  previous_version?: number;
  /** total size of the files of the current version, in bytes */
  size: number;
  file_count: number;
  /** handle of the person it was first published for, empty once that person is deleted */
  created_by: string;
  created_at: string;
  updated_at: string;
  /** who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public */
  access: string;
  /** when the access ends, after which the Window opens to its Room only; absent when it does not end */
  access_ends_at?: string;
  /** when the Window went to the trash; absent when it is not in it */
  trashed_at?: string;
  /** what the pages of the current version do, as declared when it was published; absent for a Window published before declarations */
  declaration?: Declaration;
  /** when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn */
  withdrawn?: Withdrawal;
  /** for a search, a few words of its pages around the words found; absent when its pages hold none */
  excerpt?: string;
  /** the current version, then the previous one if there is one */
  versions: Version[];
  /** the files of the current version */
  files: FileInfo[];
  /** how many share links open the Window to someone new, besides its access */
  share_links: number;
  /** its share links, the latest first, without the links themselves */
  links: ShareLink[];
  /** its database, when its declaration keeps data: its tables, their columns and rows counted, and its copies */
  data?: DataInfo;
}

export interface WindowGetInput {
  /** the Window's id or address, or its Room and its name as room/name */
  window: string;
  /** also return the content of each file */
  include_content?: boolean;
}

export interface WindowList {
  windows: Window[];
}

export interface WindowListInput {
  /** only the Windows of this Room, by its name */
  room?: string;
  /** only the Windows with this tag */
  tag?: string;
  /** words to search, such as q3 sales; accents and case do not count */
  query?: string;
  /** list the Windows in the trash instead */
  trashed?: boolean;
}

export interface WindowRef {
  /** the Room of the Window */
  room: string;
  /** the name of the Window */
  window: string;
}

export interface WithdrawInput {
  /** the Window, by its id or as room/name */
  window: string;
  /** why it is withdrawn, 500 characters at most, which the members of its Room read */
  reason: string;
}

export interface Withdrawal {
  at: string;
  /** why the moderators withdrew it, for the members of its Room */
  reason: string;
  /** the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did */
  by_server?: boolean;
}

export interface WithdrawnWindow {
  /** the Room of the Window */
  room: string;
  /** the name of the Window */
  window: string;
  /** why the moderators withdrew it */
  reason: string;
}

/** The operations of a Cahut, at POST /console/api/<operation> of its domain. */
export interface CahutOperations {
  /** Choose a password from a single-use activation link, and sign in. */
  activate_account: { input: ActivateInput; output: AccountSession };
  /** Add a person to a team Room, as a member or an owner. They can then open and publish its Windows. */
  add_room_member: { input: MemberInput; output: RoomDetail };
  /** Connect an AI client to this Cahut, for the person who consents, with the scope and the Rooms they grant. Returns the address that sends the browser back to the client, with its authorization code. */
  approve_authorization: { input: ApproveInput; output: Redirect };
  /** Approve a Knock of one of your AIs: the call runs now, as the AI, and what it returns comes to you. It may still fail, if things changed since the AI called. */
  approve_knock: { input: AnswerInput; output: Answer };
  /** Run a query of the declaration of a Window, for a page of that Window. */
  call_window_query: { input: PageDataInput; output: PageData };
  /** Change your e-mail address: a link goes to the new one, which replaces yours once you use the link, within 24 hours, and your old address gets a notice. An empty address removes yours at once. */
  change_email: { input: ChangeEmailInput; output: EmailChanged };
  /** Change your handle, with which you sign in. Your personal Room takes it as its name: the addresses of its Windows change, the old ones stop working, and your old handle is free for another person or Room. In an individual Cahut, your other Rooms, whose names start with your handle, take the new one too. Answer the Knocks that wait for you first. */
  change_handle: { input: ChangeHandleInput; output: User };
  /** Change your name, as the people of the Cahut see it. Your personal Room takes it too, unless you gave it a title of its own. */
  change_name: { input: ChangeNameInput; output: User };
  /** Change your password. Your other sessions end. */
  change_password: { input: ChangePasswordInput; output: PasswordChanged };
  /** Run SQLite statements on the database of a Window of your Rooms: create tables, add columns, add, change or delete rows. Cahut copies the database first; rollback_window_data goes back to the copy. */
  change_window_data: { input: ChangeDataInput; output: ChangedData };
  /** Say whether a handle can go to a new person: of the form of a handle, not reserved, and held by no person and by no Room. For the administrators of the Cahut. */
  check_handle: { input: CheckHandleInput; output: HandleCheck };
  /** Confirm a new e-mail address from the link sent to it, which makes it the address of its person. */
  confirm_email: { input: ConfirmEmailInput; output: EmailChanged };
  /** Give a person a new single-use link, valid 7 days, where they choose their password, for instance after they forgot it. It goes to their e-mail address too, when they have one and the Cahut sends e-mail. Their password works until they use the link. */
  create_activation_link: { input: ActivationLinkInput; output: Activation };
  /** Create a team Room, at <domain>/<name>/. You become its owner, and can then add members. Windows published in it are for its members. In an individual Cahut, you stay its only member, and its name starts with your handle and a hyphen. */
  create_room: { input: CreateRoomInput; output: Room };
  /** Create a service token, with which a program on this server manages the people of the Cahut through the admin API, without a Knock: it creates them, gives them their activation link, and disables them. */
  create_service_token: { input: CreateServiceTokenInput; output: NewServiceToken };
  /** Make a share link: whoever holds it opens the Window, without an account or a Key, until it ends, besides its access. It ends in 7 days unless you give an end, 30 days at most, and may open to a number of readers only. It is returned this once: give it to the person. revoke_share_link takes it back. Not for a public Window. */
  create_share_link: { input: ShareLinkInput; output: ShareLinkMade };
  /** Create a personal token that lets an AI act on behalf of a person. */
  create_token: { input: CreateTokenInput; output: NewToken };
  /** Create a person, with their personal Room, and the link where they choose their password, which goes to their e-mail address when you give one and the Cahut sends e-mail. */
  create_user: { input: CreateUserInput; output: NewUser };
  /** Delete a team Room for good, with its members and its Windows, those in its trash included, and their files: this cannot be undone. Its address becomes free. Its owners and the administrators of the Cahut delete it; a personal Room is never deleted. */
  delete_room: { input: DeleteRoomInput; output: DeletedRoom };
  /** Delete a person for good: their account, their tokens and the AIs they connected, their personal Room and the team Rooms they are the only owner of, with the Windows of these Rooms and their files. This cannot be undone. The Windows they published in other Rooms stay, without an author. Their handle becomes free. No one deletes themself, nor the last administrator of the Cahut; disable_user stops a person and keeps what is theirs. */
  delete_user: { input: UserInput; output: DeletedUser };
  /** Refuse the request of an AI client to connect to this Cahut. Returns the address that tells the client so. */
  deny_authorization: { input: DenyInput; output: Redirect };
  /** Deny a Knock of one of your AIs: the call never runs. The AI learns it from get_knock. */
  deny_knock: { input: AnswerInput; output: Knock };
  /** Check the request of an AI client to connect to this Cahut, and describe it to the person who consents: the client, where it will be sent back, the scope it asks for, and the Rooms the person may grant. */
  describe_authorization: { input: DescribeInput; output: Authorization };
  /** Say what the page of sign-in of this Cahut offers, before anyone signs in: whether it sends e-mail, and so a link to a person who forgot their password, the links to the legal pages of the Cahut, and whether the Cahut is suspended. */
  describe_sign_in: { input: DescribeSignInInput; output: SignInPage };
  /** Disable a person: their sessions end, and neither they nor their AIs can act until a person enables them again. Their Rooms and Windows stay. */
  disable_user: { input: UserInput; output: User };
  /** Disconnect an AI connected over OAuth. Its tokens stop working at once; to use it again, the person connects it again. */
  disconnect_ai: { input: DisconnectAIInput; output: Connection };
  /** Decide a report that waits, and leave its Window as it is, with a note for the moderators. For the moderators of this Cahut. */
  dismiss_report: { input: DismissInput; output: Report };
  /** Enable a disabled person again: they can sign in, and their AIs can act with their tokens. */
  enable_user: { input: UserInput; output: User };
  /** Describe the person you act for: their handle, name, role in the Cahut and Rooms. */
  get_account: { input: GetAccountInput; output: Account };
  /** Describe this Cahut: its domain, the address where its Rooms and Windows open, that of its MCP endpoint, whether it sends e-mail, the links to its legal pages, its kind, collaborative or individual, and its limits with what it holds: a call that would cross a limit is refused. */
  get_cahut: { input: GetCahutInput; output: About };
  /** Describe the invitation of a personal link: the Window, who invites, and their note. */
  get_invitation: { input: InvitationCodeInput; output: InvitationPage };
  /** Check a Knock of this Cahut, a call of an AI that waits for a person to approve it: whether it waits, was approved and run, failed, was denied, expired or was canceled. Call it to learn what became of a call that knocked. */
  get_knock: { input: KnockGetInput; output: Knock };
  /** Describe the SMTP relay of this Cahut, without its password, or say that it has none and sends no e-mail; and when a message last went out, or last failed and why. For the administrators of the Cahut. */
  get_mail_relay: { input: GetMailRelayInput; output: Delivery };
  /** Describe a Room of this Cahut: its title, its owners and members, and how many Windows it holds. */
  get_room: { input: RoomGetInput; output: RoomDetail };
  /** Describe one person of the Cahut, as list_users does, with their e-mail address. For the administrators of the Cahut. */
  get_user: { input: UserInput; output: User };
  /** Describe a Window of this Cahut and list its files, optionally with their contents. */
  get_window: { input: WindowGetInput; output: WindowDetail };
  /** Send the address of a Window by e-mail, with a note: to people of the Cahut, named by handle, who sign in to read it, or to anyone else, named by address, when the Window opens with its Key or to the public. For a Window with a Key, each address gets a personal link that opens it for them alone, until revoke_invitation takes it back; the Key itself never goes. 20 people a call, and 100 invitations a day. get_cahut says whether the Cahut sends e-mail. */
  invite_to_window: { input: InviteInput; output: Invited };
  /** List the AIs connected over OAuth, newest first, with the person each acts for, its scope and its Rooms. */
  list_connected_ais: { input: ListConnectedAIsInput; output: ConnectedAIs };
  /** List the invitations to a Window of this Cahut, the latest first: who got one, who invited, and whether it still stands. */
  list_invitations: { input: ListInvitationsInput; output: InvitationList };
  /** List the Knocks of this Cahut, those that wait first. A Knock is a call of an AI that waits for a person to approve it, since it cannot be undone or changes what a person may do. An AI lists the Knocks it made; a person, those of their AIs. */
  list_knocks: { input: KnockListInput; output: KnockList };
  /** Read the latest entries of the Logbook, newest first: all of them for an administrator of the Cahut, and for anyone else what they did and what AIs did on their behalf. */
  list_logbook: { input: ListLogbookInput; output: LogbookPage };
  /** List the public Windows of this Cahut, which anyone opens, the latest made public first, for its moderators to look at what it serves to anyone. */
  list_public_windows: { input: ListPublicInput; output: PublicWindows };
  /** List the reports about the Windows of this Cahut: those that wait for a decision, the oldest first, then the latest decided. For its moderators: its administrators and those of its server. */
  list_reports: { input: ListReportsInput; output: Reports };
  /** List your Rooms in this Cahut, your personal Room first, with your role and how many members and Windows each has. */
  list_rooms: { input: ListRoomsInput; output: RoomList };
  /** List the service tokens of the Cahut, newest first, with their last use. */
  list_service_tokens: { input: ListServiceTokensInput; output: ServiceTokenList };
  /** List personal tokens, newest first. */
  list_tokens: { input: ListTokensInput; output: TokenList };
  /** List the people of the Cahut, with their handles, names, roles and states. In an individual Cahut, for its administrators only. */
  list_users: { input: ListUsersInput; output: UserList };
  /** List the Windows in the Rooms of the person you act for in this Cahut, or in one of them, or those with a tag, most recently updated first; with query, those whose title, description, tags or pages hold its words, the best first. */
  list_windows: { input: WindowListInput; output: WindowList };
  /** Open the Window of the invitation of a personal link: a code, valid a minute, for /-/invitation, which gives a pass. */
  open_invitation: { input: InvitationCodeInput; output: InvitationOpening };
  /** Open the Window of a share link, from the page of the link: spend one of its openings, and give a pass. */
  open_share_link: { input: OpenShareLinkInput; output: ShareLinkOpening };
  /** Get the frame that shows, in the console, the Window of a report that waits, whatever its access, for four hours at most. The members of its Room read in the Logbook that a moderator opened it. For the moderators of this Cahut. */
  preview_reported_window: { input: PreviewReportedInput; output: ReportedPreview };
  /** Get the frames that show Windows of your Rooms in the console, as previews: each opens one Window, while your session lasts and four hours at most. */
  preview_windows: { input: PreviewInput; output: Previews };
  /** Publish a new Window in this Cahut: a document, deck, dashboard or site made of HTML and its files, in a Room of the person you act for, by default their personal Room. Returns the address where it opens. */
  publish_window: { input: PublishInput; output: Window };
  /** Delete a Window of this Cahut that is in the trash for good, with its versions and its files: this cannot be undone. Its address becomes free. Put the Window in the trash first, with trash_window. */
  purge_window: { input: RefInput; output: Window };
  /** Run one SELECT over the whole database of a Window of your Rooms, private columns included, and get 1,000 rows at most, marked cut beyond. get_window gives its tables and columns. */
  query_window_data: { input: QueryDataInput; output: QueriedData };
  /** Give a withdrawn Window back as it was: its access, its Key, its share links until their end. A Window that the administrators of the server withdrew is reinstated by them only. For the moderators of this Cahut. */
  reinstate_window: { input: ReinstateInput; output: Moderated };
  /** Remove a person from a team Room, or leave it yourself. They no longer open or publish its Windows. A Room keeps at least one owner. */
  remove_room_member: { input: RemoveMemberInput; output: RoomDetail };
  /** Change the name of a Window of this Cahut, the last part of its address, <domain>/<room>/<name>/. Its title, its id and its permanent address stay. The old address stops working at once, and its name is free in the Room: rename_window back to it undoes the change, while no other Window takes it. */
  rename_window: { input: RenameInput; output: Window };
  /** Report a Window of this Cahut to its moderators, from the page /-/report, without an account: its address, what is wrong with it, and an explanation. */
  report_window: { input: ReportInput; output: ReportReceived };
  /** Send the link where a person chooses a new password to the e-mail address of their account, given its handle or its address. The link works for one hour; the password works until it is used. The answer is the same whether the account exists or not. A Cahut without an e-mail relay sends none: an administrator of the Cahut gives a new link instead. */
  request_password_link: { input: RequestPasswordLinkInput; output: PasswordLinkRequested };
  /** Bring a Window of this Cahut back from the trash, with its versions. */
  restore_window: { input: RefInput; output: Window };
  /** Resume this Cahut, which suspend_cahut suspended: its Windows open again, its people sign in, and their AIs act in it as before, without connecting again. Its administrators hear of it by e-mail when it sends e-mail. For the administrators of the server. */
  resume_cahut: { input: ResumeCahutInput; output: Settings };
  /** Change the title of a Window of this Cahut, its description, its tags, or several, without publishing a new version. Its address stays. Give the old ones again to undo it. */
  retitle_window: { input: RetitleInput; output: Window };
  /** Take back an invitation to a Window of this Cahut: its personal link, and the passes it gave, no longer open the Window. The e-mail stays sent. */
  revoke_invitation: { input: RevokeInvitationInput; output: Invitation };
  /** Revoke a service token. The program that used it loses access to the admin API at once. */
  revoke_service_token: { input: RevokeServiceTokenInput; output: ServiceTokenInfo };
  /** Take back a share link to a Window of this Cahut: it no longer opens the Window, to those who opened it too. */
  revoke_share_link: { input: RevokeShareLinkInput; output: ShareLink };
  /** Revoke a personal token. The AI that used it loses access at once. */
  revoke_token: { input: RevokeTokenInput; output: TokenInfo };
  /** Go back to the previous version of a Window of this Cahut. The version you leave becomes the previous one, so calling rollback_window again returns to it. */
  rollback_window: { input: RefInput; output: Window };
  /** Put the database of a Window back as a copy has it, the latest by default: what readers wrote since is lost. The database as it was becomes a copy, so calling it again undoes it. */
  rollback_window_data: { input: RollbackDataInput; output: RolledBackData };
  /** Run the SELECT of a page of a Window over what its declaration makes readable. */
  select_window_data: { input: PageDataInput; output: PageData };
  /** Send a test message through the relay of this Cahut, and say whether the relay took it. For the administrators of the Cahut. */
  send_test_mail: { input: SendTestMailInput; output: TestSent };
  /** Set what an AI connected over OAuth reaches in this Cahut: read or write, and every Room of its person or some of them. A person sets the reach of their AIs. An AI sets its own, leaving out id: narrowing it acts at once, and widening it, such as asking for write or for another Room, waits for its person to approve. */
  set_ai_reach: { input: SetAIReachInput; output: Connection };
  /** Set the kind of this Cahut, collaborative or individual, and its limits: how many people, administrators, Rooms and Windows it may have, how many Windows each Room may hold, how many bytes it may keep, and, in an individual Cahut, how many Rooms, Windows and bytes each person may have. Give every limit: a limit left out is lifted. Lowering a limit deletes nothing: the Cahut keeps what it holds and refuses what would add to it. For the administrators of the server. */
  set_cahut_settings: { input: SetCahutSettingsInput; output: Settings };
  /** Set the links to the legal pages of this Cahut, which its host publishes elsewhere: its legal notice, terms, privacy policy and support. Each is the address of a page, which starts with https://; fr gives another address for a page in French, where it differs, which the console links when it speaks French. The console and the pages for readers link them at their foot. Give every link: a link left out is removed. For the administrators of the server. */
  set_legal_links: { input: SetLegalLinksInput; output: LegalLinks };
  /** Turn off, or on again, the e-mail notices that are neither about your account nor about a Knock, such as being added to a Room or the end of the access of a Window. */
  set_mail_notices: { input: SetMailNoticesInput; output: MailNotices };
  /** Set the SMTP relay that sends the e-mail of this Cahut, such as activation links, from an address of its own. The relay is at a public address and speaks TLS. Cahut connects to it and signs in before it keeps it. An empty host removes it: the Cahut then sends no e-mail, and its console offers nothing that would. For the administrators of the Cahut. */
  set_mail_relay: { input: SetMailRelayInput; output: Delivery };
  /** Set the address that receives each report about the Windows of this Cahut, beside its administrators, such as an abuse mailbox of its host, or remove it with an empty address. It receives them when the Cahut sends e-mail. For the administrators of the server. */
  set_moderation_address: { input: SetModerationAddressInput; output: ModerationAddressSet };
  /** Show a Window of a Room of this Cahut at the Room's address, <domain>/<room>/, in place of the page that lists its Windows, such as a landing page made for the Room. Readers who may not open that Window still get the list of the Windows they may open. An empty window goes back to the list. The Window keeps its own address and its access. */
  set_room_home: { input: SetHomeInput; output: RoomDetail };
  /** Make a person of a team Room an owner or a member of it. A Room keeps at least one owner. */
  set_room_member_role: { input: MemberInput; output: RoomDetail };
  /** Give a person an e-mail address, where their activation links and notices go, or remove it with an empty address. Their old address gets a notice. For the administrators of the Cahut. */
  set_user_email: { input: SetUserEmailInput; output: User };
  /** Make a person an administrator of the Cahut, or a member. The Cahut keeps at least one administrator. */
  set_user_role: { input: SetUserRoleInput; output: User };
  /** Set who opens a Window: author, its author alone, for themself; room, the members of its Room (the default); internal, everyone signed in to the Cahut, except in an individual Cahut; key, anyone who types its Key; or public, anyone with its address. The access may end at a date, after which the Window opens to its Room only. For key, Cahut draws a Key unless you give one, and returns it once: give it to the person. Making a Window public waits for the approval of a person. */
  set_window_access: { input: AccessInput; output: AccessSet };
  /** Sign in with a handle and a password, and open a session. */
  sign_in: { input: SignInInput; output: AccountSession };
  /** Get a single-use code, valid one minute, that signs you in to read the Windows of this Cahut at its address: the browser takes it to /-/session. */
  sign_in_to_windows: { input: SignInToWindowsInput; output: ReaderCode };
  /** End the session you signed in with, and the sessions it opened to read Windows. */
  sign_out: { input: SignOutInput; output: SignedOut };
  /** Suspend this Cahut, with a reason, after a decision of moderation or when its organization stops paying. It keeps everything: its people and their sessions, their AIs, its Rooms, its Windows and its Logbook. Until resume_cahut, it serves nothing to anyone but the administrators of the server: its Windows answer with a page that says it is suspended, and every call of a person, an AI or a service is refused. Its administrators receive the reason by e-mail when it sends e-mail. Suspending a suspended Cahut changes its reason only. For the administrators of the server. */
  suspend_cahut: { input: SuspendCahutInput; output: Settings };
  /** Put a Window of this Cahut in the trash: it no longer opens and leaves the lists, until restore_window brings it back as it was. */
  trash_window: { input: RefInput; output: Window };
  /** Change the title of a Room of this Cahut, its description, or both. Its name, and so its address, stays. Give the old title and description again to undo it. */
  update_room: { input: UpdateRoomInput; output: RoomDetail };
  /** Publish a new version of a Window of this Cahut: give all its files again, as for publish_window. Its address stays. Its current version becomes the previous one, which rollback_window goes back to; the version before that is deleted. Its database stays and runs the new migrations. Cahut calls no other system. */
  update_window: { input: UpdateInput; output: Window };
  /** Withdraw a Window of this Cahut that a report names, or a public Window, with a reason for the members of its Room. It keeps everything, and serves nothing to its readers until reinstate_window: its addresses and its share links say that it was withdrawn. It decides the reports that wait about it. Withdrawing a withdrawn Window changes its reason. For the moderators of this Cahut. */
  withdraw_window: { input: WithdrawInput; output: Moderated };
}

/** The operations of the server, at POST /console/api/server/<operation>, for its administrators. */
export interface ServerOperations {
  /** Choose a password from a single-use activation link, as an administrator of the server, and sign in. */
  activate_account: { input: ActivateInput; output: ServerSession };
  /** Declare a domain on this server, which creates the Cahut at that domain, optionally with its first administrator and the link where they choose their password. */
  add_domain: { input: AddDomainInput; output: NewCahut };
  /** Change your password, as an administrator of the server. Your other sessions end. */
  change_password: { input: ChangePasswordInput; output: PasswordChanged };
  /** Give an administrator of the server a new single-use link, valid 7 days, where they choose their password, for instance after they forgot it. Their password works until they use the link. */
  create_activation_link: { input: AdminInput; output: Activation };
  /** Create an administrator of the server, and the link where they choose their password. */
  create_admin: { input: CreateAdminInput; output: NewAdmin };
  /** Give the link to the first start of a Cahut that has nobody yet, single-use and valid for 24 hours, where its first person chooses their handle and password. It replaces the previous link. */
  create_setup_link: { input: CreateSetupLinkInput; output: SetupLink };
  /** Disable an administrator of the server: their sessions and their link end, and they cannot sign in until another administrator enables them again. The server keeps at least one administrator who is not disabled. */
  disable_admin: { input: AdminInput; output: Admin };
  /** Enable a disabled administrator of the server again: they can sign in with their password, or with a new link. */
  enable_admin: { input: AdminInput; output: Admin };
  /** Say whether a link to the first start still works at this domain. */
  get_setup: { input: SetupInput; output: Setup };
  /** Say which version of Cahut runs, the newest release, and how the last check and the last update went. */
  get_updates: { input: UpdatesInput; output: Updates };
  /** List the administrators of the server. */
  list_admins: { input: ListAdminsInput; output: AdminList };
  /** List the Cahuts of this server, their domains, those that are suspended, and how many reports wait in each. */
  list_domains: { input: ListDomainsInput; output: DomainList };
  /** Read the latest entries of the server's Logbook, newest first. */
  list_logbook: { input: ListLogbookInput; output: LogbookPage };
  /** Change the name of an administrator of the server, yours or another's. Their handle, with which they sign in, stays. */
  rename_admin: { input: RenameAdminInput; output: Admin };
  /** Move a Cahut to another domain, or from the IP address of the server to a domain. Its people, Rooms, Windows and Logbook stay; the links that name its old address stop working, and AIs connect again at the new one. */
  rename_domain: { input: RenameDomainInput; output: Info };
  /** Ask the updater of the server to check for new releases, or to install a release newer than the one running. It runs as root within a minute, checks the signature, and goes back to the version running if the new one fails. */
  request_update: { input: RequestUpdateInput; output: RequestedUpdate };
  /** Use a link to the first start: create the first person of the Cahut, its administrator, and the administrator of the server if it has none, with the same handle and password; then sign the person in. */
  set_up: { input: SetUpInput; output: SetUp };
  /** Sign in as an administrator of the server, with a handle and a password, and open a session at this domain. */
  sign_in: { input: SignInInput; output: ServerSession };
  /** End the session you signed in with, as an administrator of the server. */
  sign_out: { input: SignOutInput; output: SignedOut };
}

/** The refusals that the console says in the language of its person, by code, each with its values. */
export interface Refusals {
  /** this account is disabled: an administrator of the Cahut can enable it */
  account_disabled: null;
  /** this activation link is unknown, used or expired: ask an administrator of the Cahut for a new one */
  activation_link_gone: null;
  /** the Cahut needs an administrator: make another person one first */
  cahut_needs_an_admin: null;
  /** this Cahut is suspended by the administrators of its server: it keeps everything it holds, and nothing in it can be read or changed until they resume it. Ask the administrators of the Cahut, or, for an AI, tell the person you act for */
  cahut_suspended: null;
  /** the description holds a control character: write it in plain text */
  description_control: null;
  /** the description has {length} characters: use at most {limit}, a sentence or two */
  description_too_long: DescriptionLength;
  /** {email} is already your e-mail address */
  email_already_yours: TakenEmail;
  /** this link is unknown, used or expired: change your address again from your account */
  email_link_gone: null;
  /** {email} is already the e-mail address of another person of this Cahut */
  email_taken: TakenEmail;
  /** handle {handle}: use 2 to 32 lowercase letters, digits and hyphens, starting with a letter and ending with a letter or a digit */
  handle_invalid: Handle;
  /** Knocks wait for you: answer them before you change your handle, since they name the addresses of your personal Room */
  handle_knocks_waiting: null;
  /** handle {handle} is reserved: it names a page of every Cahut */
  handle_reserved: Handle;
  /** your Room {room} would be named {name}, which another Room holds or which is too long: choose another handle */
  handle_room_cannot_move: MovedRoom;
  /** handle {handle} is taken, by a person or a Room */
  handle_taken: Handle;
  /** {room}/{window} opens to its author alone, so nobody else signs in to read it, and it has no Key: make a share link instead, with create_share_link, and max_openings 1 for a single reader */
  invitation_author_only: At;
  /** this invitation does not open the Window any more: it was taken back, it expired, or the Window changed its Key or its access. Ask the person who invited you for a new one */
  invitation_gone: null;
  /** {room}/{window} opens to the people of the Cahut only: invite people of the Cahut by handle, or set its access to key first, so that each address gets a personal link */
  invite_address_internal: At;
  /** {room}/{window} opens to its Room only: invite people of the Cahut by handle, or set its access to key first, so that each address gets a personal link */
  invite_address_room_only: At;
  /** to: {to} is not an e-mail address: give one address alone, such as jane@example.com */
  invite_bad_address: Addressee;
  /** {to}: in this Cahut, which is individual, invite by e-mail address, to a Window with a Key or public */
  invite_by_address: Addressee;
  /** {handle} is disabled, and cannot read the Window */
  invite_disabled: Invitee;
  /** no person {to} in this Cahut: invite someone else by e-mail address */
  invite_no_person: Addressee;
  /** {handle} is not in the Room {room}, and {room}/{window} opens to its Room only: add them to the Room, or set its access to internal or key first */
  invite_not_in_room: Outsider;
  /** the Key starts or ends with a space */
  key_has_spaces: null;
  /** the Key has more than {limit} characters */
  key_too_long: Limit;
  /** the Key has {length} characters: give it {limit} at least, or let Cahut draw one */
  key_too_short: Length;
  /** {link}: give {page} too, the address of the page for every other language; the address in French takes its place in French only */
  legal_link_in_french_only: LinkInFrenchOnly;
  /** {link}: give the address of a page, which starts with https://, such as https://example.org/terms, or leave it out to remove the link */
  legal_link_invalid: LegalLink;
  /** name is empty */
  name_empty: null;
  /** name is longer than 200 bytes */
  name_too_long: null;
  /** this Cahut sends no e-mail: an administrator of the Cahut sets its relay first, on the page E-mail of the console */
  no_mail: null;
  /** this Cahut sends no e-mail: give them the address of the Window yourself, with its Key if it has one */
  no_mail_to_invite: null;
  /** this Cahut sends no e-mail: an administrator of the Cahut can give you a new link */
  no_password_link: null;
  /** the password has {length} characters: use at most {limit} */
  password_too_long: Length;
  /** the password has {length} characters: use at least {limit} */
  password_too_short: Length;
  /** report {report} is decided already: list_reports gives those that wait */
  report_decided: ReportRef;
  /** the note holds more than {limit} characters: keep it shorter */
  report_note_too_long: NoteLimit;
  /** name {name}: use 1 to 63 lowercase letters, digits and hyphens, starting and ending with a letter or a digit */
  room_name_invalid: NewName;
  /** name {name} is reserved: it names a page of every Cahut */
  room_name_reserved: NewName;
  /** name {name} is taken, by a Room or a person: choose another */
  room_name_taken: NewName;
  /** name {name} is not available: choose another */
  room_name_unavailable: NewName;
  /** you already have a Room {name}: choose another name */
  room_name_yours: NewName;
  /** the Room {room} needs an owner: make another person one first */
  room_needs_an_owner: NoOwner;
  /** this account is disabled: another administrator of the server can enable it */
  server_account_disabled: null;
  /** this activation link is unknown, used or expired: ask another administrator of the server for a new one, or run cahut admin link on the server */
  server_activation_link_gone: null;
  /** invalid or ended session */
  session_ended: null;
  /** the Cahut of {domain} has people already: sign in at its console */
  setup_done: Started;
  /** this link to the first start does not work any more: it was used, it expired, or it is for another address; cahut setup, on the server, gives a new one */
  setup_link_gone: null;
  /** ends_at {ends_at} is past, and a share link ends between now and {latest}: give an end in between, or none for 7 days */
  share_link_end_past: LinkEnd;
  /** share link {link} has ended already: it opens nothing more */
  share_link_ended: LinkRef;
  /** label: write it on one line, in {limit} characters at most */
  share_link_label: LinkLabel;
  /** max_openings {given}: give a number from 1 to {limit}, or none for a link that opens to any number of readers */
  share_link_openings: LinkOpenings;
  /** {room}/{window} is public: its permanent address opens it to anyone, so share that address rather than a link */
  share_link_public: At;
  /** ends_at {ends_at}: a share link lasts 30 days at most, until {latest}: give an end before it, or none for 7 days */
  share_link_too_long: LinkEnd;
  /** tag {tag}: use 1 to 32 letters, digits and hyphens, such as q3-review */
  tag_invalid: GivenTag;
  /** {count} tags: give a Window {limit} at most */
  tags_too_many: TagCount;
  /** title is empty: give one */
  title_empty: null;
  /** title is longer than 200 bytes: shorten it */
  title_too_long: null;
  /** too many attempts from this address: wait a minute */
  too_many_attempts_from_address: null;
  /** {sent} invitations went out for {handle} in the last 24 hours, and {limit} is the most: send the rest later */
  too_many_invitations: Sent;
  /** too many refused sign-ins for this handle: try again in {minutes} minutes */
  too_many_refused_sign_ins: Wait;
  /** {room}/{window} has {limit} share links that stand, the most: take one back before you make another */
  too_many_share_links: LinksOn;
  /** {made} share links were made for {handle} in the last 24 hours, and {limit} is the most: make the rest later */
  too_many_share_links_today: LinksMade;
  /** too many sign-ins at once: try again in a few seconds */
  too_many_sign_ins_at_once: null;
  /** only the author of {room}/{window}, {author}, keeps it for themself */
  window_author_only: ByAuthor;
  /** name {name}: use 1 to 63 lowercase letters, digits and hyphens, starting and ending with a letter or a digit */
  window_name_invalid: GivenName;
  /** name {name} is taken in the Room {room}, by a Window or one in its trash: choose another */
  window_name_taken: NameInRoom;
  /** {room}/{window} was withdrawn by the moderators of this Cahut: {reason}. It serves nothing, and changes nothing, until they reinstate it, which only they can do: tell the person you act for, or ask the administrators of the Cahut */
  window_withdrawn: WithdrawnWindow;
  /** the administrators of the server withdrew {room}/{window}: only they change its reason or reinstate it */
  window_withdrawn_by_host: WindowRef;
  /** {room}/{window} has no author since the person it was published for was deleted: it opens to its Room, its Cahut, a Key or the public, and to no author alone */
  window_without_author: At;
  /** give the reason of the withdrawal, in {limit} characters at most: the members of the Room of the Window read it */
  withdrawal_reason_invalid: ReasonLimit;
  /** the current password is wrong */
  wrong_current_password: null;
  /** wrong handle or password */
  wrong_handle_or_password: null;
}

/** The share of a limit, in percent, from which the console warns that it is close, as the e-mails of the Cahut do. */
export const nearLimit = 90;
