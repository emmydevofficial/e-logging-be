package middleware

import (
	"github.com/gofiber/fiber/v2"
	"e-logging-app/internal/db"
)

func DeviceFingerprintMiddleware(deviceRepo db.DeviceRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fingerprint := c.Get("X-Device-ID")
		if fingerprint == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Missing X-Device-ID header",
			})
		}

		device, err := deviceRepo.GetDeviceByFingerprint(c.Context(), fingerprint)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid device fingerprint",
			})
		}

		c.Locals("device_id", device.ID)
		return c.Next()
	}
}