package contract

func NewContractWallet(walletName string, address []byte) *ContractWallet {

	wallet := &ContractWallet{ }
	wallet.Name = walletName
	wallet.Address = address

	// For testing purpose, add 100 coins for each of the wallet
	defaultAsset := &ContractAsset{ AssetType:ContractAssetType_Coin, AssetId:0, Amount:100 }
	wallet.Assets = []*ContractAsset { defaultAsset }

	return wallet
}

func (this *ContractWallet) GetAsset(assetType ContractAssetType, assetId uint32) *ContractAsset {

	if this == nil || this.Assets == nil {
		return nil
	}

	for _, asset := range this.Assets {

		if asset.AssetType == assetType && asset.AssetId == assetId {
			return asset
		}
	}

	return nil
}

func (this *ContractWallet) HasEnough(asset *ContractAsset) bool {

	existingAsset := this.GetAsset(asset.AssetType, asset.AssetId)
	return existingAsset != nil && existingAsset.Amount >= asset.Amount
}

func (this *ContractWallet) AddAsset(newAsset *ContractAsset) bool {

	if this == nil || newAsset == nil {
		return false
	}

	if this.Assets == nil {
		this.Assets = make([]*ContractAsset, 0)
	}

	for _, asset := range this.Assets {

		if asset.AssetType == newAsset.AssetType && asset.AssetId == newAsset.AssetId {
			asset.Amount = asset.Amount + newAsset.Amount
			return true
		}
	}

	this.Assets = append(this.Assets, newAsset)
	return true
}

func (this *ContractWallet) RemoveAsset(asset *ContractAsset) bool {

	if this == nil || asset == nil {
		return false
	}

	if this.Assets == nil {
		this.Assets = make([]*ContractAsset, 0)
	}

	for _, asse := range this.Assets {

		if asse.AssetType == asset.AssetType && asse.AssetId == asset.AssetId {

			if asse.Amount >= asset.Amount {
				asse.Amount = asse.Amount - asset.Amount
				return true
			} else {
				return false
			}
		}
	}

	return false
}
