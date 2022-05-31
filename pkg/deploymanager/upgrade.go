package deploymanager

func (d *DeployManager) BeforeUpgradeCleanup(odfCatalogImage, subscriptionChannel string) error {
	err := d.UndeployODFWithOLM(odfCatalogImage, subscriptionChannel)
	if err != nil {
		return err
	}

	return nil
}
func (d *DeployManager) BeforeUpgradeSetup(upgradeFromodfCatalogImage, upgradeFromsubscriptionChannel string) error {
	err := d.DeployODFWithOLM(upgradeFromodfCatalogImage, upgradeFromsubscriptionChannel)
	if err != nil {
		return err
	}

	csvNames, err := d.GetCSVNames()
	if err != nil {
		return err
	}

	err = d.CheckAllCsvs(csvNames)
	if err != nil {
		return err
	}

	// Creating a storagecluster which we will check after upgrade
	err = d.CreateStorageSystem()
	if err != nil {
		return err
	}
	err = d.CheckStorageSystemCondition()
	if err != nil {
		return err
	}
	err = d.CheckStorageSystemOwnerRef()
	if err != nil {
		return err
	}

	return nil
}

func (d *DeployManager) UpgradeODFwithOLM(odfCatalogImage, subscriptionChannel string) error {
	olmResources := d.GetOlmResources(odfCatalogImage, subscriptionChannel)
	err := d.UpdateOlmResources(olmResources)
	if err != nil {
		return err
	}
	return nil
}

func (d *DeployManager) CheckStorageSystem() error {
	err := d.CheckStorageSystemCondition()
	if err != nil {
		return err
	}
	err = d.CheckStorageSystemOwnerRef()
	if err != nil {
		return err
	}

	return nil
}

func (d *DeployManager) AfterUpgradeCleanup() error {
	err := d.DeleteStorageSystemAndWait()
	if err != nil {
		return err
	}

	return nil
}
