interface AuthProvider {
  isAuthenticated: boolean;
  isAdmin: boolean;
  username: null | string;
  signin(username: string, isAdmin: boolean): Promise<void>;
  signout(): Promise<void>;
}

export const authProvider: AuthProvider = {
  isAuthenticated: false,
  isAdmin: false,
  username: null,
  async signin(username: string, isAdmin: boolean) {
    authProvider.isAuthenticated = true;
    authProvider.isAdmin = isAdmin;
    authProvider.username = username;
  },
  async signout() {
    authProvider.isAuthenticated = false;
    authProvider.isAdmin = false;
    authProvider.username = "";
  },
};
