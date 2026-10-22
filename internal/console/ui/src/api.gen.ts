// Generated from the Go types of the registries by make console-types: do not edit.
// See docs/adr/0014-console-http-api.md.

export interface About {
  domain: string;
  /** the address of the Cahut, under which its Rooms and Windows open */
  url?: string;
  /** the address of its MCP endpoint, where AI clients connect */
  mcp?: string;
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

export interface ChangePasswordInput {
  /** the current password */
  password: string;
  /** the new password, of 15 characters at least */
  new_password: string;
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
  token_id?: number;
  /** the AI, as its token and its software name it */
  client?: string;
  /** the adapter the call came through: cli, mcp */
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

export type ListDomainsInput = Record<string, never>;

export interface ListInput {
  /** only the Windows of this Room, by its name */
  room?: string;
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

export interface RemoveMemberInput {
  /** the Room's name */
  room: string;
  /** handle of the person; yourself to leave the Room */
  user: string;
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
  /** how many Windows the Room holds */
  windows: number;
  created_at: string;
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
  /** how many Windows the Room holds */
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

export interface Window {
  /** the Window's name, last part of its address */
  name: string;
  title: string;
  kind: string;
  /** where the Window opens */
  url: string;
  /** the name of its Room, the part of its address before its name */
  room: string;
  version: number;
  /** total size of the files, in bytes */
  size: number;
  file_count: number;
  /** handle of the person it was published for */
  created_by: string;
  created_at: string;
  updated_at: string;
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
  version: number;
  /** total size of the files, in bytes */
  size: number;
  file_count: number;
  /** handle of the person it was published for */
  created_by: string;
  created_at: string;
  updated_at: string;
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
  /** Disable a person: their sessions end, and neither they nor their AIs can act until a person enables them again. Their Rooms and Windows stay. */
  disable_user: { input: UserInput; output: User };
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
  /** Revoke a personal token. The AI that used it loses access at once. */
  revoke_token: { input: RevokeTokenInput; output: TokenInfo };
  /** Make a person of a team Room an owner or a member of it. A Room keeps at least one owner. */
  set_room_member_role: { input: MemberInput; output: RoomDetail };
  /** Make a person an administrator of the Cahut, or a member. The Cahut keeps at least one administrator. */
  set_user_role: { input: SetUserRoleInput; output: User };
  /** Sign in with a handle and a password, and open a session. */
  sign_in: { input: SignInInput; output: AccountSession };
  /** End the session you signed in with. */
  sign_out: { input: SignOutInput; output: SignedOut };
  /** Change the title of a Room. Its name, and so its address, stays. */
  update_room: { input: UpdateRoomInput; output: RoomDetail };
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
