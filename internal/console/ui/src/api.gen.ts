// Generated from the Go types of the registries by make console-types: do not edit.
// See docs/adr/0014-console-http-api.md.

export interface About {
  domain: string;
  /** the address of the Cahut, under which its Rooms and Windows open */
  url?: string;
  /** the address of its MCP endpoint, where AI clients connect */
  mcp?: string;
}

export interface AccessInput {
  /** the Window's address, or its Room and its name as room/name */
  window: string;
  /** room, internal, key or public */
  access: string;
  /** when the access ends, such as 2026-12-31T18:00:00Z; after it, the Window opens to its Room only. It does not end when absent */
  ends_at?: string;
  /** for key: the Key to set, of 10 to 256 characters. When empty, the Window keeps its Key, or Cahut draws one */
  key?: string;
  /** for key: draw a new Key; the old one no longer opens the Window */
  new_key?: boolean;
}

export interface AccessSet {
  /** the Window's name, last part of its address */
  name: string;
  title: string;
  kind: string;
  /** where the Window opens */
  url: string;
  /** the name of its Room, the part of its address before its name */
  room: string;
  /** the number of the version that readers see */
  version: number;
  /** the number of the version that rollback_window goes back to; absent when there is none */
  previous_version?: number;
  /** total size of the files of the current version, in bytes */
  size: number;
  file_count: number;
  /** handle of the person it was first published for */
  created_by: string;
  created_at: string;
  updated_at: string;
  /** who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public */
  access: string;
  /** when the access ends, after which the Window opens to its Room only; absent when it does not end */
  access_ends_at?: string;
  /** when the Window went to the trash; absent when it is not in it */
  trashed_at?: string;
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
}

export interface AddDomainInput {
  /** the domain of the new Cahut, such as cahut.example.com */
  domain: string;
  /** the first administrator of the Cahut, a person of it */
  admin?: FirstAdmin;
}

export interface Admin {
  handle: string;
  name: string;
  /** pending until the administrator chooses a password, then active */
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

export interface ApproveInput {
  request: Request;
  /** read, or write; at most what the client asks for */
  scope: string;
  /** every Room of the person, including those they join later */
  all_rooms?: boolean;
  /** the Rooms granted, by name, when all_rooms is false */
  rooms?: string[];
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

export interface ChangePasswordInput {
  /** the current password */
  password: string;
  /** the new password, of 15 characters at least */
  new_password: string;
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
  /** name of the Room, the first part of its address; from the title when empty */
  name?: string;
  /** handle of its first owner; you when empty. Only administrators of the Cahut name someone else */
  owner?: string;
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
}

export interface DenyInput {
  request: Request;
}

export interface DescribeInput {
  request: Request;
}

export interface DisconnectAIInput {
  /** the connection, from list_connected_ais */
  id: number;
}

export interface DomainList {
  cahuts: Info[];
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
  /** the adapter the call came through: cli, api, mcp or oauth */
  via: string;
  /** what the operation acted on */
  target?: string;
  detail?: string;
  /** ok, refused or failed */
  outcome: string;
  error?: string;
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

export interface Info {
  id: number;
  domain: string;
  created_at: string;
}

export type ListAdminsInput = Record<string, never>;

export interface ListConnectedAIsInput {
  /** only the AIs of this person; for a person who is not an administrator, their own */
  user?: string;
}

export type ListDomainsInput = Record<string, never>;

export interface ListInput {
  /** only the Windows of this Room, by its name */
  room?: string;
  /** list the Windows in the trash instead */
  trashed?: boolean;
}

export interface ListLogbookInput {
  /** how many entries, 50 by default, at most 1000 */
  limit?: number;
}

export interface ListRoomsInput {
  /** every Room of the Cahut, for an administrator; your Rooms when false */
  all?: boolean;
}

export interface ListTokensInput {
  /** only the tokens of this person; for a person who is not an administrator, their own tokens */
  user?: string;
}

export type ListUsersInput = Record<string, never>;

export interface LogbookPage {
  entries: Entry[];
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

export interface NewAdmin {
  handle: string;
  name: string;
  /** pending until the administrator chooses a password, then active */
  status: string;
  created_at: string;
  /** the link where the administrator chooses their password */
  activation: Activation;
}

export interface NewCahut {
  id: number;
  domain: string;
  created_at: string;
  /** the first administrator of the Cahut, with their activation link */
  admin?: NewUser;
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
  /** the link where the person chooses their password; an AI never gets it */
  activation?: Activation;
}

export type PasswordChanged = Record<string, never>;

export interface Person {
  handle: string;
  name: string;
  /** owner or member of the Room */
  role: string;
}

export interface PublishInput {
  /** title of the Window, also used in its address */
  title: string;
  /** one of document, deck, dashboard or site */
  kind: string;
  /** name of the Room to publish in; the personal Room of the person you act for when empty */
  room?: string;
  /** the files of the Window; index.html at the root is the page that opens. Link them with relative paths, such as css/deck.css */
  files: File[];
}

export interface ReaderCode {
  /** send the browser to <domain>/-/session?code=<code>, on the port of the Windows */
  code: string;
  expires_at: string;
}

export interface Redirect {
  /** send the browser to this address */
  redirect: string;
}

export interface RefInput {
  /** the Window's address, or its Room and its name as room/name */
  window: string;
}

export interface RemoveMemberInput {
  /** the Room's name */
  room: string;
  /** handle of the person; yourself to leave the Room */
  user: string;
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

export interface RevokeTokenInput {
  id: number;
}

export interface Room {
  /** the Room's name, first part of its address */
  name: string;
  title: string;
  /** personal or team */
  kind: string;
  /** your role in the Room, owner or member; empty when you are not in it */
  role?: string;
  /** how many people are in the Room */
  members: number;
  /** how many Windows the Room holds, out of the trash */
  windows: number;
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
  /** personal or team */
  kind: string;
  /** your role in the Room, owner or member; empty when you are not in it */
  role?: string;
  /** how many people are in the Room */
  members: number;
  /** how many Windows the Room holds, out of the trash */
  windows: number;
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

export interface ServerSession {
  /** send it in the Authorization header: Bearer <token> */
  token: string;
  /** the domain where the session holds */
  domain: string;
  /** when the session ends at the latest; it ends sooner after 7 days without use */
  expires_at: string;
  admin: Admin;
}

export interface SetUserRoleInput {
  /** handle of the person */
  user: string;
  /** admin or member */
  role: string;
}

export interface SignInInput {
  handle: string;
  password: string;
}

export type SignInToWindowsInput = Record<string, never>;

export type SignOutInput = Record<string, never>;

export type SignedOut = Record<string, never>;

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
  /** the Window's address, or its Room and its name as room/name */
  window: string;
  /** a new title; the title stays when empty. The address does not change */
  title?: string;
  /** every file of the new version, as for publish_window; files left out are not in it */
  files: File[];
}

export interface UpdateRoomInput {
  /** the Room's name */
  room: string;
  /** the new title */
  title: string;
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
  /** handle of the person it was published for */
  published_by: string;
  published_at: string;
}

export interface Window {
  /** the Window's name, last part of its address */
  name: string;
  title: string;
  kind: string;
  /** where the Window opens */
  url: string;
  /** the name of its Room, the part of its address before its name */
  room: string;
  /** the number of the version that readers see */
  version: number;
  /** the number of the version that rollback_window goes back to; absent when there is none */
  previous_version?: number;
  /** total size of the files of the current version, in bytes */
  size: number;
  file_count: number;
  /** handle of the person it was first published for */
  created_by: string;
  created_at: string;
  updated_at: string;
  /** who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public */
  access: string;
  /** when the access ends, after which the Window opens to its Room only; absent when it does not end */
  access_ends_at?: string;
  /** when the Window went to the trash; absent when it is not in it */
  trashed_at?: string;
}

export interface WindowDetail {
  /** the Window's name, last part of its address */
  name: string;
  title: string;
  kind: string;
  /** where the Window opens */
  url: string;
  /** the name of its Room, the part of its address before its name */
  room: string;
  /** the number of the version that readers see */
  version: number;
  /** the number of the version that rollback_window goes back to; absent when there is none */
  previous_version?: number;
  /** total size of the files of the current version, in bytes */
  size: number;
  file_count: number;
  /** handle of the person it was first published for */
  created_by: string;
  created_at: string;
  updated_at: string;
  /** who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public */
  access: string;
  /** when the access ends, after which the Window opens to its Room only; absent when it does not end */
  access_ends_at?: string;
  /** when the Window went to the trash; absent when it is not in it */
  trashed_at?: string;
  /** the current version, then the previous one if there is one */
  versions: Version[];
  /** the files of the current version */
  files: FileInfo[];
}

export interface WindowGetInput {
  /** the Window's address, or its Room and its name as room/name */
  window: string;
  /** also return the content of each file */
  include_content?: boolean;
}

export interface WindowList {
  windows: Window[];
}

/** The operations of a Cahut, at POST /api/<operation> on the console's port of its domain. */
export interface CahutOperations {
  /** Choose a password from a single-use activation link, and sign in. */
  activate_account: { input: ActivateInput; output: AccountSession };
  /** Add a person to a team Room, as a member or an owner. They can then open and publish its Windows. */
  add_room_member: { input: MemberInput; output: RoomDetail };
  /** Connect an AI client to this Cahut, for the person who consents, with the scope and the Rooms they grant. Returns the address that sends the browser back to the client, with its authorization code. */
  approve_authorization: { input: ApproveInput; output: Redirect };
  /** Change your password. Your other sessions end. */
  change_password: { input: ChangePasswordInput; output: PasswordChanged };
  /** Give a person a new single-use link, valid 7 days, where they choose their password, for instance after they forgot it. Their password works until they use the link. */
  create_activation_link: { input: UserInput; output: Activation };
  /** Create a team Room, at <domain>/<name>/. You become its owner, and can then add members. Windows published in it are for its members. */
  create_room: { input: CreateRoomInput; output: Room };
  /** Create a personal token that lets an AI act on behalf of a person. */
  create_token: { input: CreateTokenInput; output: NewToken };
  /** Create a person, with their personal Room, and the link where they choose their password. */
  create_user: { input: CreateUserInput; output: NewUser };
  /** Refuse the request of an AI client to connect to this Cahut. Returns the address that tells the client so. */
  deny_authorization: { input: DenyInput; output: Redirect };
  /** Check the request of an AI client to connect to this Cahut, and describe it to the person who consents: the client, where it will be sent back, the scope it asks for, and the Rooms the person may grant. */
  describe_authorization: { input: DescribeInput; output: Authorization };
  /** Disable a person: their sessions end, and neither they nor their AIs can act until a person enables them again. Their Rooms and Windows stay. */
  disable_user: { input: UserInput; output: User };
  /** Disconnect an AI connected over OAuth. Its tokens stop working at once; to use it again, the person connects it again. */
  disconnect_ai: { input: DisconnectAIInput; output: Connection };
  /** Enable a disabled person again: they can sign in, and their AIs can act with their tokens. */
  enable_user: { input: UserInput; output: User };
  /** Describe the person you act for: their handle, name, role in the Cahut and Rooms. */
  get_account: { input: GetAccountInput; output: Account };
  /** Describe this Cahut: its domain, the address where its Rooms and Windows open, and that of its MCP endpoint. */
  get_cahut: { input: GetCahutInput; output: About };
  /** Describe a Room: its title, its owners and members, and how many Windows it holds. */
  get_room: { input: RoomGetInput; output: RoomDetail };
  /** Describe a Window and list its files, optionally with their contents. */
  get_window: { input: WindowGetInput; output: WindowDetail };
  /** List the AIs connected over OAuth, newest first, with the person each acts for, its scope and its Rooms. */
  list_connected_ais: { input: ListConnectedAIsInput; output: ConnectedAIs };
  /** Read the latest entries of the Logbook, newest first: all of them for an administrator of the Cahut, and for anyone else what they did and what AIs did on their behalf. */
  list_logbook: { input: ListLogbookInput; output: LogbookPage };
  /** List your Rooms, your personal Room first, with your role and how many members and Windows each has. */
  list_rooms: { input: ListRoomsInput; output: RoomList };
  /** List personal tokens, newest first. */
  list_tokens: { input: ListTokensInput; output: TokenList };
  /** List the people of the Cahut, with their handles, names, roles and states. */
  list_users: { input: ListUsersInput; output: UserList };
  /** List the Windows in the Rooms of the person you act for, or in one of them, most recently updated first. */
  list_windows: { input: ListInput; output: WindowList };
  /** Publish a new Window: a document, deck, dashboard or site made of HTML and its files, in a Room of the person you act for, by default their personal Room. Returns the address where it opens. */
  publish_window: { input: PublishInput; output: Window };
  /** Remove a person from a team Room, or leave it yourself. They no longer open or publish its Windows. A Room keeps at least one owner. */
  remove_room_member: { input: RemoveMemberInput; output: RoomDetail };
  /** Bring a Window back from the trash, with its versions. */
  restore_window: { input: RefInput; output: Window };
  /** Revoke a personal token. The AI that used it loses access at once. */
  revoke_token: { input: RevokeTokenInput; output: TokenInfo };
  /** Go back to the previous version of a Window. The version you leave becomes the previous one, so calling rollback_window again returns to it. */
  rollback_window: { input: RefInput; output: Window };
  /** Make a person of a team Room an owner or a member of it. A Room keeps at least one owner. */
  set_room_member_role: { input: MemberInput; output: RoomDetail };
  /** Make a person an administrator of the Cahut, or a member. The Cahut keeps at least one administrator. */
  set_user_role: { input: SetUserRoleInput; output: User };
  /** Set who opens a Window: room, the members of its Room (the default); internal, everyone signed in to the Cahut; key, anyone who types its Key; or public, anyone with its address. The access may end at a date, after which the Window opens to its Room only. For key, Cahut draws a Key unless you give one, and returns it once: give it to the person. Making a Window public waits for the approval of a person. */
  set_window_access: { input: AccessInput; output: AccessSet };
  /** Sign in with a handle and a password, and open a session. */
  sign_in: { input: SignInInput; output: AccountSession };
  /** Get a single-use code, valid one minute, that signs you in to read the Windows of this Cahut at its address: the browser takes it to /-/session. */
  sign_in_to_windows: { input: SignInToWindowsInput; output: ReaderCode };
  /** End the session you signed in with. */
  sign_out: { input: SignOutInput; output: SignedOut };
  /** Put a Window in the trash: it no longer opens and leaves the lists, until restore_window brings it back as it was. */
  trash_window: { input: RefInput; output: Window };
  /** Change the title of a Room. Its name, and so its address, stays. */
  update_room: { input: UpdateRoomInput; output: RoomDetail };
  /** Publish a new version of a Window: give all its files again, as for publish_window. Its address stays. Its current version becomes the previous one, which rollback_window goes back to; the version before that is deleted. */
  update_window: { input: UpdateInput; output: Window };
}

/** The operations of the server, at POST /api/server/<operation>, for its administrators. */
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
  /** List the administrators of the server. */
  list_admins: { input: ListAdminsInput; output: AdminList };
  /** List the Cahuts of this server and their domains. */
  list_domains: { input: ListDomainsInput; output: DomainList };
  /** Read the latest entries of the server's Logbook, newest first. */
  list_logbook: { input: ListLogbookInput; output: LogbookPage };
  /** Sign in as an administrator of the server, with a handle and a password, and open a session at this domain. */
  sign_in: { input: SignInInput; output: ServerSession };
  /** End the session you signed in with, as an administrator of the server. */
  sign_out: { input: SignOutInput; output: SignedOut };
}
