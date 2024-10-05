interface AuthProvider {
  isAuthenticated: boolean;
  username: null | string;
  signin(username: string): Promise<void>;
  signout(): Promise<void>;
}

export const authProvider: AuthProvider = {
  isAuthenticated: false,
  username: null,
  async signin(username: string) {
    authProvider.isAuthenticated = true;
    authProvider.username = username;
  },
  async signout() {
    authProvider.isAuthenticated = false;
    authProvider.username = "";
  },
};
