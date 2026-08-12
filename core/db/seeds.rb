# Seeds are for local development only — never run against staging or production.
unless Rails.env.development?
  puts "WARN: Seeding is just for development!"
else
  def find_or_create_identity(email, staff: false)
    identity = Identity.find_or_initialize_by(email: email)
    identity.staff = staff if staff
    identity.save!
    identity
  end

  def create_account(name:, owner:, description: nil, owner_name: nil)
    return if owner.accounts.where(personal: false, name: name).exists?

    Account.create_with_owner(
      account: { name: name, description: description || name },
      owner: { name: owner_name || owner.full_name, identity: owner }
    )
  end

  # Main seed workflow
  yuler = find_or_create_identity("yuler@example.com", staff: true)
  create_account(
    name: "Yuler Company",
    description: "Yuler's company",
    owner: yuler,
    owner_name: "Yuler Doe"
  )

  puts "OK: Seeded yuler identity and Yuler Company account (development only)."
end
