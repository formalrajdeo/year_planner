import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { AuthPage } from "./AuthPage";

afterEach(cleanup);

describe("AuthPage", () => {
  it("collects registration fields and submits them to the auth handler", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn().mockResolvedValue(undefined);
    render(
      <AuthPage
        onSubmit={onSubmit}
        onRequestPasswordReset={vi.fn()}
        onResetPassword={vi.fn()}
        error=""
      />,
    );

    await user.click(screen.getByRole("tab", { name: "Create account" }));
    await user.type(screen.getByRole("textbox", { name: "Your name" }), "Maya Chen");
    await user.type(screen.getByRole("textbox", { name: "Email address" }), "maya@example.com");
    await user.type(screen.getByPlaceholderText("At least 8 characters"), "planner-password");
    await user.click(screen.getByRole("button", { name: "Create your account" }));

    expect(onSubmit).toHaveBeenCalledWith("register", {
      displayName: "Maya Chen",
      email: "maya@example.com",
      password: "planner-password",
    });
  });

  it("requests password recovery without leaving the auth page", async () => {
    const user = userEvent.setup();
    const onRequestPasswordReset = vi.fn().mockResolvedValue({
      message: "If an account exists, instructions will be sent.",
    });
    render(
      <AuthPage
        onSubmit={vi.fn()}
        onRequestPasswordReset={onRequestPasswordReset}
        onResetPassword={vi.fn()}
        error=""
      />,
    );

    await user.click(screen.getByRole("button", { name: "Forgot password?" }));
    await user.type(screen.getByRole("textbox", { name: "Email address" }), "maya@example.com");
    await user.click(screen.getByRole("button", { name: "Send reset instructions" }));

    expect(onRequestPasswordReset).toHaveBeenCalledWith("maya@example.com");
    const status = await screen.findByRole("status");
    expect(status.textContent).toContain("If an account exists");
  });

  it("submits a matching password with the reset token", async () => {
    const user = userEvent.setup();
    const onResetPassword = vi.fn().mockResolvedValue(undefined);
    render(
      <AuthPage
        onSubmit={vi.fn()}
        onRequestPasswordReset={vi.fn()}
        onResetPassword={onResetPassword}
        passwordResetToken={"a".repeat(64)}
        error=""
      />,
    );

    await user.type(screen.getByPlaceholderText("At least 8 characters"), "new-password");
    await user.type(
      screen.getByPlaceholderText("Enter your new password again"),
      "new-password",
    );
    await user.click(screen.getByRole("button", { name: "Update password" }));

    expect(onResetPassword).toHaveBeenCalledWith("a".repeat(64), "new-password");
    const status = await screen.findByRole("status");
    expect(status.textContent).toContain("password has been updated");
  });
});
