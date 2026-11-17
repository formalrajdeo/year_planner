import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { AuthPage } from "./AuthPage";

describe("AuthPage", () => {
  it("collects registration fields and submits them to the auth handler", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn().mockResolvedValue(undefined);
    render(<AuthPage onSubmit={onSubmit} error="" />);

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
});
